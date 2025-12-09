-- +goose Up
-- +goose StatementBegin

-- Наполнение таблицы violations начальными данными
INSERT INTO violations (code, description) VALUES
('SPEED', 'Превышение скорости'),
('RED_LIGHT', 'Проезд на красный свет'),
('DRUNK_DRIVING', 'Вождение в нетрезвом виде'),
('SEAT_BELT', 'Нарушение правил использования ремня безопасности'),
('PHONE_USE', 'Использование телефона за рулем'),
('WRONG_LANE', 'Движение по встречной полосе'),
('STOP_SIGN', 'Проезд на знак Стоп'),
('NO_INSURANCE', 'Отсутствие страховки на автомобиль'),
('DANGEROUS_MANEUVER', 'Опасный маневр на дороге'),
('PARKING_VIOLATION', 'Нарушение правил парковки'),
('ILLEGAL_OVERTAKE', 'Незаконный обгон'),
('NO_LICENSE', 'Отсутствие водительского удостоверения'),
('EXPIRED_LICENSE', 'Истек срок действия водительского удостоверения'),
('NO_SEATBELT_CHILD', 'Отсутствие детского удерживающего устройства'),
('OVERSIZED_LOAD', 'Перевозка грузов с превышением допустимых габаритов'),
('UNSAFE_VEHICLE', 'Ненадлежащий технический осмотр транспортного средства'),
('ILLEGAL_U_TURN', 'Незаконный разворот'),
('FAILED_SIGNAL', 'Несоблюдение сигналов светофора'),
('STOPPING_ON_HIGHWAY', 'Остановка на автомагистрали в неположенном месте'),
('OTHER', 'Другое нарушение');

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Очистка таблицы violations (удаляем добавленные записи)
DELETE FROM violations
WHERE code IN (
    'SPEED','RED_LIGHT','DRUNK_DRIVING','SEAT_BELT','PHONE_USE',
    'WRONG_LANE','STOP_SIGN','NO_INSURANCE','DANGEROUS_MANEUVER','PARKING_VIOLATION',
    'ILLEGAL_OVERTAKE','NO_LICENSE','EXPIRED_LICENSE','NO_SEATBELT_CHILD','OVERSIZED_LOAD',
    'UNSAFE_VEHICLE','ILLEGAL_U_TURN','FAILED_SIGNAL','STOPPING_ON_HIGHWAY','OTHER'
);

-- +goose StatementEnd
