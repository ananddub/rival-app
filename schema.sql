-- Users table (assuming it already exists, just showing the coins column addition)
ALTER TABLE users
ADD COLUMN coins INT DEFAULT 0;

-- Coin Transactions table
CREATE TABLE coin_transactions (
    id INT PRIMARY KEY AUTO_INCREMENT,
    user_id INT NOT NULL,
    amount INT NOT NULL, -- positive for earn, negative for spend
    type ENUM('earn', 'spend') NOT NULL,
    reason VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);
