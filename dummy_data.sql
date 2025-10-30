-- Insert dummy users with hashed passwords (bcrypt hash for "password123")
INSERT INTO public.users (full_name, email, phone_number, password_hash, dob, is_phone_verified, is_email_verified, sign_type, role) VALUES
('John Doe', 'john@example.com', '+919876543210', '$2b$10$K7L/8Y75aO7O2fOxlXXXXeJ4kP8/f1AZXvCXXXXXXXXXXXXXXXXXXX', '1990-05-15', true, true, 'email', 'user'),
('Jane Smith', 'jane@example.com', '+919876543211', '$2b$10$K7L/8Y75aO7O2fOxlXXXXeJ4kP8/f1AZXvCXXXXXXXXXXXXXXXXXXX', '1992-08-22', true, true, 'email', 'user'),
('Admin User', 'admin@example.com', '+919876543212', '$2b$10$K7L/8Y75aO7O2fOxlXXXXeJ4kP8/f1AZXvCXXXXXXXXXXXXXXXXXXX', '1985-12-10', true, true, 'email', 'admin'),
('Mike Johnson', 'mike@example.com', '+919876543213', '$2b$10$K7L/8Y75aO7O2fOxlXXXXeJ4kP8/f1AZXvCXXXXXXXXXXXXXXXXXXX', '1988-03-18', false, true, 'email', 'user'),
('Sarah Wilson', 'sarah@example.com', '+919876543214', '$2b$10$K7L/8Y75aO7O2fOxlXXXXeJ4kP8/f1AZXvCXXXXXXXXXXXXXXXXXXX', '1995-11-07', true, false, 'phone', 'user');

-- Insert wallets for users
INSERT INTO public.wallets (user_id, balance, coins, currency) VALUES
(1, 5000.50, 100, 'INR'),
(2, 2500.75, 50, 'INR'),
(3, 10000.00, 200, 'INR'),
(4, 1200.25, 25, 'INR'),
(5, 750.00, 15, 'INR');

-- Insert transactions
INSERT INTO public.transactions (user_id, wallet_id, title, description, amount, type, icon) VALUES
(1, 1, 'Salary Credit', 'Monthly salary deposit', 50000.00, 'credit', '💰'),
(1, 1, 'Grocery Shopping', 'Weekly grocery expenses', -2500.50, 'debit', '🛒'),
(2, 2, 'Freelance Payment', 'Web development project', 15000.00, 'credit', '💻'),
(2, 2, 'Electricity Bill', 'Monthly electricity payment', -1200.25, 'debit', '⚡'),
(3, 3, 'Investment Return', 'Stock market gains', 25000.00, 'credit', '📈'),
(4, 4, 'Food Delivery', 'Dinner order', -450.75, 'debit', '🍕'),
(5, 5, 'Gift Received', 'Birthday gift money', 2000.00, 'credit', '🎁');

-- Insert activities
INSERT INTO public.activities (user_id, action, details, category, icon) VALUES
(1, 'Account Created', 'User successfully registered', 'account', '👤'),
(1, 'Transaction Made', 'Grocery shopping payment', 'transaction', '💳'),
(2, 'Profile Updated', 'Changed phone number', 'profile', '✏️'),
(2, 'Money Received', 'Freelance payment credited', 'transaction', '💰'),
(3, 'Login', 'Admin logged into dashboard', 'security', '🔐'),
(4, 'Wallet Topped Up', 'Added money to wallet', 'wallet', '💵'),
(5, 'Password Changed', 'Security password updated', 'security', '🔒');

-- Insert notifications
INSERT INTO public.notifications (user_id, title, message, type, is_read) VALUES
(1, 'Welcome!', 'Welcome to our platform. Complete your profile to get started.', 'info', false),
(1, 'Transaction Alert', 'Your grocery payment of ₹2500.50 was successful.', 'success', true),
(2, 'Payment Received', 'You received ₹15000 for your freelance project.', 'success', false),
(3, 'Admin Alert', 'New user registrations increased by 25% this week.', 'info', true),
(4, 'Low Balance', 'Your wallet balance is running low. Consider adding funds.', 'warning', false),
(5, 'Security Update', 'Your password was successfully changed.', 'info', true);

-- Insert JWT tokens (sample active tokens)
INSERT INTO public.jwt_tokens (user_id, token_hash, expires_at) VALUES
(1, 'abc123def456ghi789jkl012mno345pqr678stu901vwx234yz', '2025-10-26 01:36:44'),
(2, 'def456ghi789jkl012mno345pqr678stu901vwx234yz567abc', '2025-10-25 15:30:00'),
(3, 'ghi789jkl012mno345pqr678stu901vwx234yz567abc123def', '2025-10-27 09:15:30');


-- Insert dummy rewards
INSERT INTO rewards (user_id, title, description, points, type, status) VALUES
                                                                            (1, 'Daily Login', 'Login for 7 consecutive days', 100, 'daily', 'available'),
                                                                            (1, 'First Purchase', 'Make your first purchase', 50, 'milestone', 'claimed'),
                                                                            (2, 'Referral Master', 'Refer 5 friends', 200, 'referral', 'available'),
                                                                            (3, 'Profile Complete', 'Complete your profile 100%', 75, 'profile', 'available');

