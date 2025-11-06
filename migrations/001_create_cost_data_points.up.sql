-- Enable UUID extension for generating UUIDs
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create cost_data_points table
CREATE TABLE IF NOT EXISTS cost_data_points (
    id UUID NOT NULL DEFAULT uuid_generate_v4(),
    category VARCHAR(100) NOT NULL,
    sub_category VARCHAR(100),
    item_name VARCHAR(255) NOT NULL,
    price DECIMAL(10, 2) NOT NULL,
    min_price DECIMAL(10, 2),
    max_price DECIMAL(10, 2),
    median_price DECIMAL(10, 2),
    sample_size INTEGER DEFAULT 1,
    location JSONB NOT NULL,
    recorded_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    valid_from TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    valid_to TIMESTAMPTZ,
    source VARCHAR(255) NOT NULL,
    source_url TEXT,
    confidence REAL DEFAULT 1.0 CHECK (confidence >= 0 AND confidence <= 1),
    unit VARCHAR(50) NOT NULL DEFAULT 'AED',
    tags TEXT[],
    attributes JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (id, recorded_at)
);

-- Create indexes for common queries
CREATE INDEX idx_cost_data_points_category ON cost_data_points(category);
CREATE INDEX idx_cost_data_points_sub_category ON cost_data_points(sub_category);
CREATE INDEX idx_cost_data_points_item_name ON cost_data_points(item_name);
CREATE INDEX idx_cost_data_points_recorded_at ON cost_data_points(recorded_at DESC);
CREATE INDEX idx_cost_data_points_location ON cost_data_points USING GIN(location);
CREATE INDEX idx_cost_data_points_tags ON cost_data_points USING GIN(tags);
CREATE INDEX idx_cost_data_points_attributes ON cost_data_points USING GIN(attributes);

-- Create composite index for location queries
CREATE INDEX idx_cost_data_points_location_emirate ON cost_data_points((location->>'emirate'));
CREATE INDEX idx_cost_data_points_location_city ON cost_data_points((location->>'city'));

-- Convert to TimescaleDB hypertable
SELECT create_hypertable('cost_data_points', 'recorded_at',
    chunk_time_interval => INTERVAL '1 month',
    if_not_exists => TRUE
);

-- Create updated_at trigger function
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create trigger to auto-update updated_at
CREATE TRIGGER update_cost_data_points_updated_at
    BEFORE UPDATE ON cost_data_points
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
