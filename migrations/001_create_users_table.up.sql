-- Create users table for Tribute application
CREATE TABLE IF NOT EXISTS users (
    user_id BIGINT PRIMARY KEY,
    earned NUMERIC(10, 2) DEFAULT 0.00,
    is_verified BOOLEAN DEFAULT FALSE,
    is_sub_published BOOLEAN DEFAULT FALSE,
    is_onboarded BOOLEAN DEFAULT FALSE
);

-- Create index on user_id for faster lookups
CREATE INDEX IF NOT EXISTS idx_users_user_id ON users(user_id); 