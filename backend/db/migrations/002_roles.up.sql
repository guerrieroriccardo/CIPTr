-- Add role column, migrate is_admin data, drop is_admin, make password_hash nullable.
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'users' AND column_name = 'is_admin'
    ) THEN
        ALTER TABLE users ADD COLUMN role TEXT NOT NULL DEFAULT 'technician';
        UPDATE users SET role = 'admin' WHERE is_admin = true;
        UPDATE users SET role = 'technician' WHERE is_admin = false;
        ALTER TABLE users DROP COLUMN is_admin;
    END IF;

    -- Make password_hash nullable (guest accounts have no password).
    ALTER TABLE users ALTER COLUMN password_hash DROP NOT NULL;
END
$$;
