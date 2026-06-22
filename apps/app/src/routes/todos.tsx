import { Button } from "@repo/design-system/components/ui/button";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@repo/design-system/components/ui/card";
import { Checkbox } from "@repo/design-system/components/ui/checkbox";
import { Input } from "@repo/design-system/components/ui/input";
import { toast } from "@repo/design-system/components/ui/sonner";
import { createFileRoute } from "@tanstack/react-router";
import { Trash2Icon } from "lucide-react";
import { type FormEvent, useState } from "react";
import { RequireAuth } from "../components/require-auth";
import { apiUrl } from "../features";
import {
  type Todo,
  useCreateTodo,
  useDeleteTodo,
  useTodos,
  useUpdateTodo,
} from "../lib/todos";

export const Route = createFileRoute("/todos")({
  component: TodosPage,
});

function TodosPage() {
  return (
    <RequireAuth>
      <Todos />
    </RequireAuth>
  );
}

function Todos() {
  const { data: todos, isLoading, error } = useTodos();
  const createTodo = useCreateTodo();
  const [title, setTitle] = useState("");

  const remaining = todos?.filter((t) => !t.completed).length ?? 0;

  function handleAdd(e: FormEvent) {
    e.preventDefault();
    const trimmed = title.trim();
    if (!trimmed) {
      return;
    }
    createTodo.mutate(trimmed, {
      onSuccess: () => setTitle(""),
      onError: (err) => toast.error(err.message),
    });
  }

  return (
    <div className="mx-auto w-full max-w-2xl space-y-6 p-6">
      <div className="flex items-center justify-between">
        <h1 className="font-bold text-2xl">Todos</h1>
        {todos ? (
          <span className="text-muted-foreground text-sm">
            {remaining} remaining
          </span>
        ) : null}
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Add a todo</CardTitle>
        </CardHeader>
        <CardContent>
          <form className="flex gap-2" onSubmit={handleAdd}>
            <Input
              maxLength={50}
              onChange={(e) => setTitle(e.target.value)}
              placeholder="What needs doing?"
              value={title}
            />
            <Button disabled={!title.trim() || createTodo.isPending} type="submit">
              {createTodo.isPending ? "Adding…" : "Add"}
            </Button>
          </form>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Your list</CardTitle>
        </CardHeader>
        <CardContent>
          {isLoading ? <p className="text-sm">Loading todos…</p> : null}
          {error ? (
            <p className="text-destructive text-sm">
              Could not reach the API. Is it running on {apiUrl}?
            </p>
          ) : null}
          {todos && todos.length === 0 ? (
            <p className="text-muted-foreground text-sm">
              Nothing yet — add your first todo above.
            </p>
          ) : null}
          {todos && todos.length > 0 ? (
            <ul className="divide-y">
              {todos.map((todo) => (
                <TodoRow key={todo.id} todo={todo} />
              ))}
            </ul>
          ) : null}
        </CardContent>
      </Card>
    </div>
  );
}

function TodoRow({ todo }: { todo: Todo }) {
  const updateTodo = useUpdateTodo();
  const deleteTodo = useDeleteTodo();

  function toggle(completed: boolean) {
    updateTodo.mutate(
      { ...todo, completed },
      { onError: (err) => toast.error(err.message) }
    );
  }

  function remove() {
    deleteTodo.mutate(todo.id, {
      onSuccess: () => toast.success("Todo deleted"),
      onError: (err) => toast.error(err.message),
    });
  }

  return (
    <li className="flex items-center gap-3 py-3">
      <Checkbox
        checked={todo.completed}
        disabled={updateTodo.isPending}
        id={`todo-${todo.id}`}
        onCheckedChange={(value) => toggle(value === true)}
      />
      <label
        className={
          todo.completed
            ? "flex-1 text-muted-foreground text-sm line-through"
            : "flex-1 text-sm"
        }
        htmlFor={`todo-${todo.id}`}
      >
        {todo.title}
      </label>
      <Button
        aria-label="Delete todo"
        disabled={deleteTodo.isPending}
        onClick={remove}
        size="icon"
        variant="ghost"
      >
        <Trash2Icon className="size-4" />
      </Button>
    </li>
  );
}
