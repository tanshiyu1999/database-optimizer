-- San Francisco Fire Calls Table Schema
DROP TABLE IF EXISTS fire_calls;

CREATE TABLE fire_calls (
  call_number VARCHAR(20),
  unit_id VARCHAR(10),
  incident_number BIGINT,
  call_type VARCHAR(100),
  call_date VARCHAR(20),
  watch_date VARCHAR(20),
  call_final_disposition VARCHAR(50),
  available_dt_tm VARCHAR(50),
  address VARCHAR(200),
  city VARCHAR(50),
  zipcode VARCHAR(10),
  battalion VARCHAR(10),
  station_area VARCHAR(10),
  box VARCHAR(10),
  original_priority INTEGER,
  priority INTEGER,
  final_priority INTEGER,
  als_unit BOOLEAN,
  call_type_group VARCHAR(100),
  num_alarms INTEGER,
  unit_type VARCHAR(50),
  unit_sequence_in_call_dispatch INTEGER,
  fire_prevention_district VARCHAR(10),
  supervisor_district VARCHAR(10),
  neighborhood VARCHAR(100),
  location VARCHAR(100),
  row_id VARCHAR(50) PRIMARY KEY,
  delay NUMERIC(15, 8)
);

-- Create indexes for common query patterns
-- CREATE INDEX idx_call_type ON fire_calls(call_type);
-- CREATE INDEX idx_call_date ON fire_calls(call_date);
-- CREATE INDEX idx_neighborhood ON fire_calls(neighborhood);
-- CREATE INDEX idx_unit_type ON fire_calls(unit_type);
