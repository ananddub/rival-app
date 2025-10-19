
-- Insert dummy rewards
INSERT INTO rewards (user_id, title, description, points, type, status) VALUES
(1, 'Daily Login', 'Login for 7 consecutive days', 100, 'daily', 'available'),
(1, 'First Purchase', 'Make your first purchase', 50, 'milestone', 'claimed'),
(2, 'Referral Master', 'Refer 5 friends', 200, 'referral', 'available'),
(3, 'Profile Complete', 'Complete your profile 100%', 75, 'profile', 'available');

