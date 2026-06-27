import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useApiClient } from "./api";

/**
 * Todo is the normalised shape the UI consumes. The generated client types
 * every field as optional (swag emits them that way), so we coerce to concrete
 * values once, here, and let the rest of the app rely on a clean type.
 */
export interface Todo {
  completed: boolean;
  createdAt: string;
  id: string;
  title: string;
  updatedAt: string;
}

const TODOS_KEY = ["todos"] as const;

/** useTodos fetches the todo list. */
export function useTodos() {
  const api = useApiClient();
  return useQuery({
    queryKey: TODOS_KEY,
    queryFn: async (): Promise<Todo[]> => {
      const { data, error } = await api.GET("/todos");
      if (error || !data) {
        throw new Error("Failed to load todos");
      }
      return data.map(normalize);
    },
  });
}

/** useCreateTodo adds a new todo, then refreshes the list. */
export function useCreateTodo() {
  const api = useApiClient();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (title: string): Promise<Todo> => {
      const { data, error } = await api.POST("/todos", { body: { title } });
      if (error || !data) {
        throw new Error("Failed to create todo");
      }
      return normalize(data);
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: TODOS_KEY }),
  });
}

/** useUpdateTodo edits a todo's title and/or completion state. */
export function useUpdateTodo() {
  const api = useApiClient();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (todo: Todo): Promise<Todo> => {
      const { data, error } = await api.PUT("/todos/{id}", {
        params: { path: { id: todo.id } },
        body: { title: todo.title, completed: todo.completed },
      });
      if (error || !data) {
        throw new Error("Failed to update todo");
      }
      return normalize(data);
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: TODOS_KEY }),
  });
}

/** useDeleteTodo removes a todo. */
export function useDeleteTodo() {
  const api = useApiClient();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (id: string): Promise<void> => {
      const { error } = await api.DELETE("/todos/{id}", {
        params: { path: { id } },
      });
      if (error) {
        throw new Error("Failed to delete todo");
      }
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: TODOS_KEY }),
  });
}

/** normalize coerces the all-optional generated type into a concrete Todo. */
function normalize(raw: {
  id?: string;
  title?: string;
  completed?: boolean;
  createdAt?: string;
  updatedAt?: string;
}): Todo {
  return {
    id: raw.id ?? "",
    title: raw.title ?? "",
    completed: raw.completed ?? false,
    createdAt: raw.createdAt ?? "",
    updatedAt: raw.updatedAt ?? "",
  };
}
