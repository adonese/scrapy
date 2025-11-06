-- Drop trigger and function
DROP TRIGGER IF EXISTS update_cost_data_points_updated_at ON cost_data_points;
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop table (this will also drop all indexes and hypertable)
DROP TABLE IF EXISTS cost_data_points CASCADE;

-- Note: We don't drop the uuid-ossp extension as it might be used by other tables
