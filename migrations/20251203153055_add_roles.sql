-- +goose Up
-- +goose StatementBegin
INSERT INTO roles (name, description) VALUES
('admin', 'Полный доступ к системе, можно смотреть абсолютно все данные'),
('inspector', 'Выписывание штрафов, а также фиксирование нарушений, занесение протоколов'),
('analyst', 'Аналитик, который имеет возможность формирования отчёта и просмотра статистики');

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM roles;
-- +goose StatementEnd
