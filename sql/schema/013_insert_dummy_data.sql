-- +goose Up

-- Insert dummy users (using IDs that don't conflict)
INSERT INTO users (full_name, email, phone_number, password_hash, dob, is_phone_verified, is_email_verified, sign_type, role) VALUES
('John Doe', 'john@example.com', '+1234567890', '$2a$10$N9qo8uLOickgx2ZMRZoMye.IjPeOXANBjH/JYW19HsxD', '1990-01-15', true, true, 'email', 'user'),
('Jane Smith', 'jane@example.com', '+1234567891', '$2a$10$N9qo8uLOickgx2ZMRZoMye.IjPeOXANBjH/JYW19HsxD', '1992-05-20', true, false, 'email', 'user'),
('Mike Johnson', 'mike@example.com', '+1234567892', '$2a$10$N9qo8uLOickgx2ZMRZoMye.IjPeOXANBjH/JYW19HsxD', '1988-12-10', false, true, 'email', 'user'),
('Sarah Wilson', 'sarah@example.com', '+1234567893', '$2a$10$N9qo8uLOickgx2ZMRZoMye.IjPeOXANBjH/JYW19HsxD', '1995-03-25', true, true, 'email', 'admin'),
('Alex Brown', 'alex@example.com', '+1234567894', '$2a$10$N9qo8uLOickgx2ZMRZoMye.IjPeOXANBjH/JYW19HsxD', '1993-08-14', true, true, 'email', 'user')
ON CONFLICT (email) DO NOTHING;

-- Insert dummy wallets
INSERT INTO wallets (user_id, balance, coins, currency) 
SELECT u.id, balance, coins, currency FROM (
    VALUES 
    ('john@example.com', 1500.50, 250, 'INR'),
    ('jane@example.com', 2300.75, 180, 'USD'),
    ('mike@example.com', 890.25, 320, 'INR'),
    ('sarah@example.com', 5000.00, 500, 'EUR'),
    ('alex@example.com', 1200.80, 150, 'INR')
) AS w(email, balance, coins, currency)
JOIN users u ON u.email = w.email
ON CONFLICT (user_id) DO NOTHING;

-- Insert dummy transactions
INSERT INTO transactions (user_id, wallet_id, title, description, amount, type, icon) 
SELECT u.id, w.id, title, description, amount, type, icon FROM (
    VALUES 
    ('john@example.com', 'Money Added', 'Added money to wallet', 500.00, 'credit', '💰'),
    ('john@example.com', 'Purchase', 'Bought premium package', -50.00, 'debit', '🛒'),
    ('jane@example.com', 'Transfer Received', 'Money from friend', 200.00, 'credit', '📥'),
    ('jane@example.com', 'Withdrawal', 'ATM withdrawal', -100.00, 'debit', '🏧'),
    ('mike@example.com', 'Salary', 'Monthly salary credit', 800.00, 'credit', '💼')
) AS t(email, title, description, amount, type, icon)
JOIN users u ON u.email = t.email
JOIN wallets w ON w.user_id = u.id;

-- Insert dummy coin transactions
INSERT INTO coin_transactions (user_id, coins, type, reason) 
SELECT u.id, coins, type, reason FROM (
    VALUES
    ('john@example.com', 50, 'earn', 'Daily login bonus'),
    ('john@example.com', -20, 'spend', 'Purchased avatar'),
    ('jane@example.com', 100, 'earn', 'Completed profile'),
    ('jane@example.com', -30, 'spend', 'Bought premium theme'),
    ('mike@example.com', 75, 'earn', 'Referral bonus')
) AS c(email, coins, type, reason)
JOIN users u ON u.email = c.email;

-- Insert dummy notifications
INSERT INTO notifications (user_id, title, message, type, is_read) 
SELECT u.id, title, message, type, is_read FROM (
    VALUES
    ('john@example.com', 'Welcome!', 'Welcome to Rival App! Start earning coins now.', 'welcome', false),
    ('john@example.com', 'Coins Earned', 'You earned 50 coins for daily login!', 'reward', true),
    ('jane@example.com', 'Profile Complete', 'Your profile is now 100% complete. Earn bonus coins!', 'achievement', false),
    ('mike@example.com', 'Payment Received', 'You received money from friend', 'transaction', true),
    ('sarah@example.com', 'New Feature', 'Check out our new reward system!', 'announcement', false)
) AS n(email, title, message, type, is_read)
JOIN users u ON u.email = n.email;

-- +goose Down
-- Clean up dummy data
DELETE FROM notifications WHERE user_id IN (SELECT id FROM users WHERE email LIKE '%@example.com');
DELETE FROM coin_transactions WHERE user_id IN (SELECT id FROM users WHERE email LIKE '%@example.com');
DELETE FROM transactions WHERE user_id IN (SELECT id FROM users WHERE email LIKE '%@example.com');
DELETE FROM wallets WHERE user_id IN (SELECT id FROM users WHERE email LIKE '%@example.com');
DELETE FROM users WHERE email LIKE '%@example.com';
