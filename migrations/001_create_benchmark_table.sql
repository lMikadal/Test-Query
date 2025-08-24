-- +goose Up
-- Create benchmark_data table for testing
CREATE TABLE benchmark_data (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(150) NOT NULL UNIQUE,
    age INTEGER CHECK (age > 0 AND age < 150),
    city VARCHAR(100),
    country VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    random_number INTEGER,
    description TEXT
);

-- Create indexes for better query performance
CREATE INDEX idx_benchmark_data_email ON benchmark_data(email);
CREATE INDEX idx_benchmark_data_city ON benchmark_data(city);
CREATE INDEX idx_benchmark_data_created_at ON benchmark_data(created_at);
CREATE INDEX idx_benchmark_data_age ON benchmark_data(age);
CREATE INDEX idx_benchmark_data_country_city ON benchmark_data(country, city);

-- Create function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create trigger to automatically update updated_at
CREATE TRIGGER update_benchmark_data_updated_at
    BEFORE UPDATE ON benchmark_data
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- +goose Down
-- Drop the table and related objects
DROP TRIGGER IF EXISTS update_benchmark_data_updated_at ON benchmark_data;
DROP FUNCTION IF EXISTS update_updated_at_column();
DROP TABLE IF EXISTS benchmark_data;
