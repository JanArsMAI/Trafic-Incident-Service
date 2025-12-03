-- +goose Up
-- +goose StatementBegin

CREATE INDEX idx_users_role_id ON users(role_id);

CREATE INDEX idx_drivers_full_name ON drivers(full_name);

CREATE INDEX idx_vehicles_owner_driver_id ON vehicles(owner_driver_id);

CREATE INDEX idx_inspectors_department ON inspectors(department);
CREATE INDEX idx_inspectors_full_name ON inspectors(full_name);

CREATE INDEX idx_accidents_date_time ON accidents(date_time);
CREATE INDEX idx_accidents_weather_id ON accidents(weather_id);
CREATE INDEX idx_accidents_inspector_id ON accidents(inspector_id);
CREATE INDEX idx_accidents_severity ON accidents(severity);

CREATE INDEX idx_participants_accident_id ON accident_participants(accident_id);
CREATE INDEX idx_participants_driver_id ON accident_participants(driver_id);
CREATE INDEX idx_participants_vehicle_id ON accident_participants(vehicle_id);

CREATE INDEX idx_violations_description ON violations USING gin (to_tsvector('russian', description));

CREATE INDEX idx_partviolations_participant_id ON participant_violations(participant_id);
CREATE INDEX idx_partviolations_violation_id ON participant_violations(violation_id);

CREATE INDEX idx_penalties_participant_id ON penalties(participant_id);
CREATE INDEX idx_penalties_issued_by ON penalties(issued_by);
CREATE INDEX idx_penalties_issue_date ON penalties(issue_date);
CREATE INDEX idx_penalties_status ON penalties(status);

CREATE INDEX idx_reports_accident_id ON reports(accident_id);
CREATE INDEX idx_reports_inspector_id ON reports(inspector_id);
CREATE INDEX idx_reports_created_at ON reports(created_at);

CREATE INDEX idx_auditlog_user_id ON audit_log(user_id);
CREATE INDEX idx_auditlog_table_name ON audit_log(table_name);
CREATE INDEX idx_auditlog_timestamp ON audit_log(timestamp);

-- Быстрый поиск неоплаченных штрафов
CREATE INDEX idx_penalties_unpaid ON penalties(issue_date) WHERE status = 'unpaid';

-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_penalties_unpaid;

DROP INDEX IF EXISTS idx_auditlog_timestamp;
DROP INDEX IF EXISTS idx_auditlog_table_name;
DROP INDEX IF EXISTS idx_auditlog_user_id;

DROP INDEX IF EXISTS idx_reports_created_at;
DROP INDEX IF EXISTS idx_reports_inspector_id;
DROP INDEX IF EXISTS idx_reports_accident_id;

DROP INDEX IF EXISTS idx_penalties_status;
DROP INDEX IF EXISTS idx_penalties_issue_date;
DROP INDEX IF EXISTS idx_penalties_issued_by;
DROP INDEX IF EXISTS idx_penalties_participant_id;

DROP INDEX IF EXISTS idx_partviolations_violation_id;
DROP INDEX IF EXISTS idx_partviolations_participant_id;

DROP INDEX IF EXISTS idx_violations_description;

DROP INDEX IF EXISTS idx_participants_vehicle_id;
DROP INDEX IF EXISTS idx_participants_driver_id;
DROP INDEX IF EXISTS idx_participants_accident_id;

DROP INDEX IF EXISTS idx_accidents_severity;
DROP INDEX IF EXISTS idx_accidents_inspector_id;
DROP INDEX IF EXISTS idx_accidents_weather_id;
DROP INDEX IF EXISTS idx_accidents_date_time;

DROP INDEX IF EXISTS idx_inspectors_full_name;
DROP INDEX IF EXISTS idx_inspectors_department;

DROP INDEX IF EXISTS idx_vehicles_owner_driver_id;

DROP INDEX IF EXISTS idx_drivers_full_name;

DROP INDEX IF EXISTS idx_users_role_id;

-- +goose StatementEnd
