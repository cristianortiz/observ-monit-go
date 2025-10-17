-- migrations/000001_create_users_table.down.sql

-- delete trigger
DROP TRIGGER IF EXISTS update_users_updated_at ON users;

-- delete function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- delete table (indexes are automatically removed)
DROP TABLE IF EXISTS users;