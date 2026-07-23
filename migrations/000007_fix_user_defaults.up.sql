UPDATE users SET login = email WHERE login IS NULL;
UPDATE users SET role = 'user' WHERE role IS NULL;
UPDATE users SET subscription = 'free' WHERE subscription IS NULL;