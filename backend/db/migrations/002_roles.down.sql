-- Revert: add is_admin back, drop role, make password_hash NOT NULL.
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'users' AND column_name = 'role'
    ) THEN
        ALTER TABLE users ADD COLUMN is_admin BOOLEAN DEFAULT false;
        UPDATE users SET is_admin = true WHERE role = 'admin';
        ALTER TABLE users DROP COLUMN role;
    END IF;

    -- Remove guest users (they have no password, can't satisfy NOT NULL).
    DELETE FROM users WHERE password_hash IS NULL;
    ALTER TABLE users ALTER COLUMN password_hash SET NOT NULL;
END
$$;
