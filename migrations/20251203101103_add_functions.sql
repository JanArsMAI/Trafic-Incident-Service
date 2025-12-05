-- +goose Up
-- +goose StatementBegin


-- Количество ДТП у водителя
CREATE OR REPLACE FUNCTION get_driver_accident_count(driver_id INT)
RETURNS INTEGER AS $$
    SELECT COUNT(*)
    FROM accident_participants
    WHERE driver_id = get_driver_accident_count.driver_id;
$$ LANGUAGE SQL;


-- Общая сумма штрафов участника ДТП
CREATE OR REPLACE FUNCTION get_total_penalties(participant_id INT)
RETURNS NUMERIC AS $$
    SELECT COALESCE(SUM(amount), 0)
    FROM penalties
    WHERE participant_id = get_total_penalties.participant_id
      AND status != 'canceled';
$$ LANGUAGE SQL;


-- Рейтинг водителя: опыт / (1 + количество ДТП)
CREATE OR REPLACE FUNCTION get_driver_rating(driver_id INT)
RETURNS NUMERIC AS $$
    SELECT ROUND(
        d.experience_years::NUMERIC / (1 + COUNT(ap.id)),
        2
    )
    FROM drivers d
    LEFT JOIN accident_participants ap ON ap.driver_id = d.id
    WHERE d.id = driver_id
    GROUP BY d.experience_years;
$$ LANGUAGE SQL;


-- Проверка виновности участника ДТП
CREATE OR REPLACE FUNCTION is_participant_guilty(part_id INT)
RETURNS BOOLEAN AS $$
    SELECT is_guilty
    FROM accident_participants
    WHERE id = part_id;
$$ LANGUAGE SQL;




-- ДТП по водителю
CREATE OR REPLACE FUNCTION get_accidents_by_driver(driver_id INT)
RETURNS TABLE(
    accident_id INT,
    date_time TIMESTAMP,
    location VARCHAR,
    severity VARCHAR,
    is_guilty BOOLEAN
) AS $$
    SELECT 
        a.id,
        a.date_time,
        a.location,
        a.severity,
        ap.is_guilty
    FROM accident_participants ap
    JOIN accidents a ON a.id = ap.accident_id
    WHERE ap.driver_id = driver_id;
$$ LANGUAGE SQL;


-- Сумма штрафов по ДТП
CREATE OR REPLACE FUNCTION get_penalty_summary(accident_id INT)
RETURNS TABLE(
    participant_id INT,
    total_amount NUMERIC,
    paid NUMERIC,
    unpaid NUMERIC
) AS $$
    SELECT 
        p.participant_id,
        SUM(p.amount) AS total_amount,
        SUM(CASE WHEN p.status = 'paid' THEN p.amount ELSE 0 END) AS paid,
        SUM(CASE WHEN p.status = 'unpaid' THEN p.amount ELSE 0 END) AS unpaid
    FROM penalties p
    WHERE p.participant_id IN (
        SELECT id
        FROM accident_participants
        WHERE accident_id = get_penalty_summary.accident_id
    )
    GROUP BY p.participant_id;
$$ LANGUAGE SQL;


-- Полный отчёт по ДТП (JSONB)
CREATE OR REPLACE FUNCTION get_full_accident_report(accident_id INT)
RETURNS TABLE(
    accident_id INT,
    date_time TIMESTAMP,
    location VARCHAR,
    severity VARCHAR,
    weather JSONB,
    inspector JSONB,
    participants JSONB
) AS $$
    SELECT 
        a.id,
        a.date_time,
        a.location,
        a.severity,
        to_jsonb(w.*) AS weather,
        to_jsonb(i.*) AS inspector,
        (
            SELECT jsonb_agg(jsonb_build_object(
                'participant_id', ap.id,
                'driver', to_jsonb(d.*),
                'vehicle', to_jsonb(v.*),
                'is_guilty', ap.is_guilty,
                'injuries', ap.injuries,
                'violations', (
                    SELECT jsonb_agg(to_jsonb(vi.*))
                    FROM participant_violations pv
                    JOIN violations vi ON pv.violation_id = vi.id
                    WHERE pv.participant_id = ap.id
                )
            ))
            FROM accident_participants ap
            JOIN drivers d ON ap.driver_id = d.id
            JOIN vehicles v ON ap.vehicle_id = v.id
            WHERE ap.accident_id = a.id
        ) AS participants
    FROM accidents a
    LEFT JOIN weather w ON a.weather_id = w.id
    LEFT JOIN inspectors i ON a.inspector_id = i.id
    WHERE a.id = accident_id;
$$ LANGUAGE SQL;


-- Нарушения одного водителя
CREATE OR REPLACE FUNCTION get_driver_violations(driver_id INT)
RETURNS TABLE(
    accident_id INT,
    violation_code VARCHAR,
    description TEXT
) AS $$
    SELECT
        ap.accident_id,
        v.code,
        v.description
    FROM accident_participants ap
    JOIN participant_violations pv ON pv.participant_id = ap.id
    JOIN violations v ON v.id = pv.violation_id
    WHERE ap.driver_id = driver_id;
$$ LANGUAGE SQL;

-- +goose StatementEnd



-- +goose Down
-- +goose StatementBegin

DROP FUNCTION IF EXISTS get_driver_accident_count(INT);
DROP FUNCTION IF EXISTS get_total_penalties(INT);
DROP FUNCTION IF EXISTS get_driver_rating(INT);
DROP FUNCTION IF EXISTS is_participant_guilty(INT);

DROP FUNCTION IF EXISTS get_accidents_by_driver(INT);
DROP FUNCTION IF EXISTS get_penalty_summary(INT);
DROP FUNCTION IF EXISTS get_full_accident_report(INT);
DROP FUNCTION IF EXISTS get_driver_violations(INT);

-- +goose StatementEnd
