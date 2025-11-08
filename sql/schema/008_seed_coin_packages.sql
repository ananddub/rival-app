-- +goose Up
INSERT INTO public.coin_packages (name, coins, price, bonus_coins) VALUES
('Starter Pack', 100, 99.00, 0),
('Bronze Pack', 500, 449.00, 50),
('Silver Pack', 1000, 849.00, 150),
('Gold Pack', 2500, 1999.00, 500),
('Platinum Pack', 5000, 3799.00, 1200),
('Diamond Pack', 10000, 6999.00, 3000);

-- +goose Down
DELETE FROM public.coin_packages;
