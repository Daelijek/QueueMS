-- Create clients table
CREATE TABLE clients (
    id SERIAL PRIMARY KEY,
    queue_id INT NOT NULL,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE, -- Email must be unique
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Index for faster lookups on queue_id
CREATE INDEX idx_clients_queue_id ON clients (queue_id);
