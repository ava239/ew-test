CREATE TABLE subscriptions (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    service_name VARCHAR(255) NOT NULL,
    price INT NOT NULL,
    user_id UUID NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE
);
CREATE INDEX service_name ON subscriptions (service_name);
CREATE INDEX user_id ON subscriptions (user_id);
CREATE INDEX start_date ON subscriptions (start_date);

INSERT INTO subscriptions VALUES
    (DEFAULT, 'Yandex Plus', 400,'60601fee-2bf1-4721-ae6f-7636e79a0cba','2025-07-01', NULL),
    (DEFAULT, 'SberPrime', 450,'60601fee-2bf1-4721-ae6f-7636e79a0cba','2025-03-01', NULL),
    (DEFAULT, 'Netflix', 1000,'60601fee-2bf1-4721-ae6f-7636e79a0cba','2025-01-01', '2025-06-01'),
    (DEFAULT, 'Spotify', 100,'9c7b87a8-b1fa-410a-9c73-92c6089b4c39','2025-03-01', NULL),
    (DEFAULT, 'Netflix', 950,'9c7b87a8-b1fa-410a-9c73-92c6089b4c39','2025-01-01', NULL),
    (DEFAULT, 'Yandex Plus', 950,'1c60d9d4-3425-424c-893e-7b670500bac0','2024-01-01', NULL);
