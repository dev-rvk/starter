-- Create "todo" table
CREATE TABLE "todo" (
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  "id" text NOT NULL,
  "title" text NOT NULL,
  "completed" boolean NULL DEFAULT false,
  PRIMARY KEY ("id")
);
