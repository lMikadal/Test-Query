-- Initialize database with sample data for benchmarking
-- This script will create a table with 300,000 rows for testing

-- Create a sample table for benchmark testing
CREATE TABLE IF NOT EXISTS benchmark_data (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(150) NOT NULL,
    age INTEGER,
    city VARCHAR(100),
    country VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    random_number INTEGER,
    description TEXT
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_benchmark_data_email ON benchmark_data(email);
CREATE INDEX IF NOT EXISTS idx_benchmark_data_city ON benchmark_data(city);
CREATE INDEX IF NOT EXISTS idx_benchmark_data_created_at ON benchmark_data(created_at);

-- Insert sample data (this will be done via application for better control)
-- The application will handle inserting 300,000 records for testing
