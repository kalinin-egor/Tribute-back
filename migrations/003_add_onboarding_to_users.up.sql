-- Add is_onboarded column to users table
ALTER TABLE users ADD COLUMN is_onboarded BOOLEAN DEFAULT FALSE; 