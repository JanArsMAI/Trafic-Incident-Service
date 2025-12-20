-- +goose Up
-- +goose StatementBegin
-- Функция аудита
CREATE OR REPLACE FUNCTION log_audit()
RETURNS TRIGGER AS $$
DECLARE
    cur_user TEXT;
    cur_user_id INTEGER;
BEGIN
    cur_user := current_setting('app.current_user_id', true);
    IF cur_user IS NULL OR cur_user = '' THEN
        cur_user_id := NULL;
    ELSE
        cur_user_id := cur_user::INTEGER;
    END IF;

    INSERT INTO audit_log(user_id, action, table_name, record_id, old_data, new_data)
    VALUES (
        cur_user_id,
        TG_OP,
        TG_TABLE_NAME,
        COALESCE(NEW.id, OLD.id),
        to_jsonb(OLD),
        to_jsonb(NEW)
    );

    IF TG_OP = 'DELETE' THEN
        RETURN OLD;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Функция автообновления updated_at
CREATE OR REPLACE FUNCTION update_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;


-- Триггер пересчёта количества ДТП у водителя
CREATE OR REPLACE FUNCTION recalc_driver_accidents()
RETURNS TRIGGER AS $$
DECLARE
    d_id INTEGER;
BEGIN
    d_id := COALESCE(NEW.driver_id, OLD.driver_id);

    UPDATE drivers
    SET total_accidents = (
        SELECT COUNT(*) FROM accident_participants ap
        WHERE ap.driver_id = d_id
    )
    WHERE drivers.id = d_id;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_driver_accidents_insert
AFTER INSERT ON accident_participants
FOR EACH ROW EXECUTE FUNCTION recalc_driver_accidents();

CREATE TRIGGER trg_driver_accidents_delete
AFTER DELETE ON accident_participants
FOR EACH ROW EXECUTE FUNCTION recalc_driver_accidents();

CREATE TRIGGER trg_driver_accidents_update
AFTER UPDATE OF driver_id ON accident_participants
FOR EACH ROW EXECUTE FUNCTION recalc_driver_accidents();

--Триггеры audit на все таблицы

CREATE TRIGGER trg_roles_audit
AFTER INSERT OR UPDATE OR DELETE ON roles
FOR EACH ROW EXECUTE FUNCTION log_audit();

CREATE TRIGGER trg_users_audit
AFTER INSERT OR UPDATE OR DELETE ON users
FOR EACH ROW EXECUTE FUNCTION log_audit();

CREATE TRIGGER trg_drivers_audit
AFTER INSERT OR UPDATE OR DELETE ON drivers
FOR EACH ROW EXECUTE FUNCTION log_audit();

CREATE TRIGGER trg_vehicles_audit
AFTER INSERT OR UPDATE OR DELETE ON vehicles
FOR EACH ROW EXECUTE FUNCTION log_audit();

CREATE TRIGGER trg_inspectors_audit
AFTER INSERT OR UPDATE OR DELETE ON inspectors
FOR EACH ROW EXECUTE FUNCTION log_audit();

CREATE TRIGGER trg_weather_audit
AFTER INSERT OR UPDATE OR DELETE ON weather
FOR EACH ROW EXECUTE FUNCTION log_audit();

CREATE TRIGGER trg_accidents_audit
AFTER INSERT OR UPDATE OR DELETE ON accidents
FOR EACH ROW EXECUTE FUNCTION log_audit();

CREATE TRIGGER trg_accident_participants_audit
AFTER INSERT OR UPDATE OR DELETE ON accident_participants
FOR EACH ROW EXECUTE FUNCTION log_audit();

CREATE TRIGGER trg_violations_audit
AFTER INSERT OR UPDATE OR DELETE ON violations
FOR EACH ROW EXECUTE FUNCTION log_audit();

CREATE TRIGGER trg_participant_violations_audit
AFTER INSERT OR UPDATE OR DELETE ON participant_violations
FOR EACH ROW EXECUTE FUNCTION log_audit();

CREATE TRIGGER trg_penalties_audit
AFTER INSERT OR UPDATE OR DELETE ON penalties
FOR EACH ROW EXECUTE FUNCTION log_audit();

CREATE TRIGGER trg_reports_audit
AFTER INSERT OR UPDATE OR DELETE ON reports
FOR EACH ROW EXECUTE FUNCTION log_audit();

--Автообновление updated_at
CREATE TRIGGER trg_users_timestamp
BEFORE UPDATE ON users
FOR EACH ROW EXECUTE FUNCTION update_timestamp();

CREATE TRIGGER trg_reports_timestamp
BEFORE UPDATE ON reports
FOR EACH ROW EXECUTE FUNCTION update_timestamp();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS trg_users_timestamp ON users;
DROP TRIGGER IF EXISTS trg_reports_timestamp ON reports;

DROP TRIGGER IF EXISTS trg_roles_audit ON roles;
DROP TRIGGER IF EXISTS trg_users_audit ON users;
DROP TRIGGER IF EXISTS trg_drivers_audit ON drivers;
DROP TRIGGER IF EXISTS trg_vehicles_audit ON vehicles;
DROP TRIGGER IF EXISTS trg_inspectors_audit ON inspectors;
DROP TRIGGER IF EXISTS trg_weather_audit ON weather;
DROP TRIGGER IF EXISTS trg_accidents_audit ON accidents;
DROP TRIGGER IF EXISTS trg_accident_participants_audit ON accident_participants;
DROP TRIGGER IF EXISTS trg_violations_audit ON violations;
DROP TRIGGER IF EXISTS trg_participant_violations_audit ON participant_violations;
DROP TRIGGER IF EXISTS trg_penalties_audit ON penalties;
DROP TRIGGER IF EXISTS trg_reports_audit ON reports;

DROP TRIGGER IF EXISTS trg_driver_accidents_insert ON accident_participants;
DROP TRIGGER IF EXISTS trg_driver_accidents_delete ON accident_participants;
DROP TRIGGER IF EXISTS trg_driver_accidents_update ON accident_participants;

DROP FUNCTION IF EXISTS recalc_driver_accidents();
ALTER TABLE drivers DROP COLUMN IF EXISTS total_accidents;

DROP FUNCTION IF EXISTS update_timestamp();
DROP FUNCTION IF EXISTS log_audit();
-- +goose StatementEnd
