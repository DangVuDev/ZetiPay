CREATE TABLE transactions (
    id SERIAL PRIMARY KEY,
    wallet_address VARCHAR(42) NOT NULL,
    tx_hash VARCHAR(66) UNIQUE NOT NULL,
    type VARCHAR(20) NOT NULL,
    amount VARCHAR(50) NOT NULL,
    to_address VARCHAR(42) NOT NULL,
    token_address VARCHAR(42),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (wallet_address) REFERENCES users(wallet_address)
);