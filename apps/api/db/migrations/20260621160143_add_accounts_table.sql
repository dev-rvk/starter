-- Create "accounts" table
CREATE TABLE "accounts" (
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  "id" text NOT NULL,
  "email" text NOT NULL,
  "password_hash" text NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "accounts_email_key" UNIQUE ("email")
);
