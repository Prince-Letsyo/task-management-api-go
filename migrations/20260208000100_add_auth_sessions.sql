-- Add refresh token version and expand TOTP secret size
ALTER TABLE "user"
  ADD COLUMN IF NOT EXISTS "refresh_token_version" integer NOT NULL DEFAULT 0;

ALTER TABLE "user"
  ALTER COLUMN "totp_secret" TYPE character varying(255);

-- Create "session" table
CREATE TABLE IF NOT EXISTS "session" (
  "id" character varying(36) NOT NULL,
  "user_id" bigint NOT NULL,
  "refresh_token_hash" character varying(128) NOT NULL,
  "user_agent" character varying(512) NULL,
  "ip_address" character varying(64) NULL,
  "last_used_at" timestamptz NULL,
  "revoked_at" timestamptz NULL,
  "created_at" timestamptz NULL DEFAULT now(),
  "updated_at" timestamptz NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_user_session" FOREIGN KEY ("user_id") REFERENCES "user" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS "idx_session_user_id" ON "session" ("user_id");
