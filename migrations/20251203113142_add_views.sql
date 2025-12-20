-- +goose Up
-- +goose StatementBegin

-- ДТП по водителям
CREATE OR REPLACE VIEW view_driver_accident_stats AS
SELECT
    d.id AS driver_id,
    d.full_name,
    d.experience_years,
    COUNT(ap.id) AS accidents_count,
    SUM(CASE WHEN ap.is_guilty THEN 1 ELSE 0 END) AS guilty_count
FROM drivers d
LEFT JOIN accident_participants ap ON ap.driver_id = d.id
GROUP BY d.id;


-- Сводка штрафов
CREATE OR REPLACE VIEW view_penalty_summary AS
SELECT
    ap.driver_id,
    d.full_name,
    COUNT(p.id) AS penalties_total,
    SUM(p.amount) AS total_amount,
    SUM(CASE WHEN p.status = 'paid' THEN p.amount ELSE 0 END) AS paid_amount,
    SUM(CASE WHEN p.status = 'unpaid' THEN p.amount ELSE 0 END) AS unpaid_amount
FROM penalties p
JOIN accident_participants ap ON p.participant_id = ap.id
JOIN drivers d ON d.id = ap.driver_id
GROUP BY ap.driver_id, d.full_name;


-- Аналитический отчёт по ДТП
-- объединение: ДТП + инспектор + погода + кол-во участников
CREATE OR REPLACE VIEW view_accident_report AS
SELECT
    a.id AS accident_id,
    a.date_time,
    a.location,
    a.severity,
    w.temperature,
    w.precipitation,
    w.visibility,
    w.road_condition,
    i.full_name AS inspector_name,
    i.department AS inspector_department,
    (
        SELECT COUNT(*) 
        FROM accident_participants ap 
        WHERE ap.accident_id = a.id
    ) AS participants_count,
    (
        SELECT COUNT(*) 
        FROM accident_participants ap 
        WHERE ap.accident_id = a.id AND ap.is_guilty = TRUE
    ) AS guilty_participants
FROM accidents a
LEFT JOIN weather w ON w.id = a.weather_id
LEFT JOIN inspectors i ON i.id = a.inspector_id;

-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin

DROP VIEW IF EXISTS view_accident_report;
DROP VIEW IF EXISTS view_penalty_summary;
DROP VIEW IF EXISTS view_driver_accident_stats;

-- +goose StatementEnd