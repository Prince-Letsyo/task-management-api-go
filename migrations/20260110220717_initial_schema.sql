-- Create "user" table
CREATE TABLE "user" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "first_name" text NOT NULL,
  "last_name" text NOT NULL,
  "username" text NOT NULL,
  "email" text NOT NULL,
  "password" character varying(256) NOT NULL,
  "email_verified" boolean NULL DEFAULT false,
  "is_admin" boolean NULL DEFAULT false,
  "is_2fa_enabled" boolean NOT NULL DEFAULT false,
  "totp_secret" character varying(32) NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_user_deleted_at" to table: "user"
CREATE INDEX "idx_user_deleted_at" ON "user" ("deleted_at");
-- Create index "idx_user_email" to table: "user"
CREATE UNIQUE INDEX "idx_user_email" ON "user" ("email");
-- Create index "idx_user_user_name" to table: "user"
CREATE UNIQUE INDEX "idx_user_user_name" ON "user" ("username");
-- Create "profile" table
CREATE TABLE "profile" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "user_id" bigint NOT NULL,
  "bio" text NULL,
  "avatar_url" text NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_user_profile" FOREIGN KEY ("user_id") REFERENCES "user" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "idx_profile_deleted_at" to table: "profile"
CREATE INDEX "idx_profile_deleted_at" ON "profile" ("deleted_at");
-- Create index "idx_profile_user_id" to table: "profile"
CREATE UNIQUE INDEX "idx_profile_user_id" ON "profile" ("user_id");
