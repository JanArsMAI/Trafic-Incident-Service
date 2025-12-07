-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS roles (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE,
    description TEXT
);

CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(100) NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    role_id INTEGER NOT NULL REFERENCES roles(id) ON DELETE RESTRICT ON UPDATE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS drivers (
    id SERIAL PRIMARY KEY,
    full_name VARCHAR(200) NOT NULL,
    date_of_birth DATE NOT NULL,
    total_accidents INTEGER NOT NULL DEFAULT 0,
    license_number VARCHAR(50) NOT NULL UNIQUE,
    license_issue_date DATE NOT NULL,
    experience_years INTEGER CHECK (experience_years >= 0),
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS vehicles (
    id SERIAL PRIMARY KEY,
    plate_number VARCHAR(20) NOT NULL UNIQUE,
    model VARCHAR(100) NOT NULL,
    year INTEGER CHECK (year >= 1950 AND year <= EXTRACT(YEAR FROM NOW())),
    vehicle_type VARCHAR(50) NOT NULL,
    owner_driver_id INTEGER REFERENCES drivers(id) ON DELETE SET NULL ON UPDATE CASCADE,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS inspectors (
    id SERIAL PRIMARY KEY,
    full_name VARCHAR(200) NOT NULL,
    badge_number VARCHAR(50) UNIQUE NOT NULL,
    department VARCHAR(100) NOT NULL,
    rank VARCHAR(100),
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS weather (
    id SERIAL PRIMARY KEY,
    temperature NUMERIC(5,2),
    precipitation VARCHAR(50),
    visibility INTEGER CHECK (visibility >= 0),
    road_condition VARCHAR(100),
    description TEXT
);

CREATE TABLE IF NOT EXISTS accidents (
    id SERIAL PRIMARY KEY,
    location VARCHAR(255) NOT NULL,
    date_time TIMESTAMP NOT NULL,
    weather_id INTEGER REFERENCES weather(id) ON DELETE SET NULL ON UPDATE CASCADE,
    inspector_id INTEGER REFERENCES inspectors(id) ON DELETE SET NULL ON UPDATE CASCADE,
    severity VARCHAR(50) CHECK (severity IN ('low','medium','high','fatal')),
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS accident_participants (
    id SERIAL PRIMARY KEY,
    accident_id INTEGER NOT NULL REFERENCES accidents(id) ON DELETE CASCADE ON UPDATE CASCADE,
    driver_id INTEGER NOT NULL REFERENCES drivers(id) ON DELETE RESTRICT ON UPDATE CASCADE,
    vehicle_id INTEGER NOT NULL REFERENCES vehicles(id) ON DELETE RESTRICT ON UPDATE CASCADE,
    is_guilty BOOLEAN NOT NULL DEFAULT FALSE,
    injuries VARCHAR(255),
    UNIQUE(accident_id, driver_id, vehicle_id)
);

CREATE TABLE IF NOT EXISTS violations (
    id SERIAL PRIMARY KEY,
    code VARCHAR(50) NOT NULL UNIQUE,
    description TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS participant_violations (
    id SERIAL PRIMARY KEY,
    participant_id INTEGER NOT NULL REFERENCES accident_participants(id) ON DELETE CASCADE,
    violation_id INTEGER NOT NULL REFERENCES violations(id) ON DELETE RESTRICT,
    UNIQUE(participant_id, violation_id)
);

CREATE TABLE IF NOT EXISTS penalties (
    id SERIAL PRIMARY KEY,
    participant_id INTEGER NOT NULL REFERENCES accident_participants(id) ON DELETE CASCADE,
    amount NUMERIC(10,2) NOT NULL CHECK (amount >= 0),
    issued_by INTEGER REFERENCES inspectors(id) ON DELETE SET NULL,
    issue_date DATE NOT NULL DEFAULT NOW(),
    status VARCHAR(50) CHECK (status IN ('unpaid','paid','canceled')) NOT NULL DEFAULT 'unpaid'
);

CREATE TABLE IF NOT EXISTS reports (
    id SERIAL PRIMARY KEY,
    accident_id INTEGER NOT NULL REFERENCES accidents(id) ON DELETE CASCADE,
    inspector_id INTEGER REFERENCES inspectors(id) ON DELETE SET NULL,
    report_text TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS audit_log (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    action VARCHAR(50) NOT NULL,
    table_name VARCHAR(100) NOT NULL,
    record_id INTEGER,
    old_data JSONB,
    new_data JSONB,
    timestamp TIMESTAMP NOT NULL DEFAULT NOW()
);

-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS audit_log;
DROP TABLE IF EXISTS reports;
DROP TABLE IF EXISTS penalties;
DROP TABLE IF EXISTS participant_violations;
DROP TABLE IF EXISTS violations;
DROP TABLE IF EXISTS accident_participants;
DROP TABLE IF EXISTS accidents;
DROP TABLE IF EXISTS weather;
DROP TABLE IF EXISTS inspectors;
DROP TABLE IF EXISTS vehicles;
DROP TABLE IF EXISTS drivers;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS roles;

-- +goose StatementEnd
