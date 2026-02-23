-- Seed cities
INSERT INTO cities (id, name, country, code) VALUES
  ('c0000001-0000-0000-0000-000000000001', 'Mumbai', 'India', 'BOM'),
  ('c0000001-0000-0000-0000-000000000002', 'Delhi', 'India', 'DEL'),
  ('c0000001-0000-0000-0000-000000000003', 'Bangalore', 'India', 'BLR'),
  ('c0000001-0000-0000-0000-000000000004', 'Chennai', 'India', 'MAA'),
  ('c0000001-0000-0000-0000-000000000005', 'Kolkata', 'India', 'CCU'),
  ('c0000001-0000-0000-0000-000000000006', 'Hyderabad', 'India', 'HYD'),
  ('c0000001-0000-0000-0000-000000000007', 'Pune', 'India', 'PNQ'),
  ('c0000001-0000-0000-0000-000000000008', 'Ahmedabad', 'India', 'AMD'),
  ('c0000001-0000-0000-0000-000000000009', 'Goa', 'India', 'GOI'),
  ('c0000001-0000-0000-0000-000000000010', 'Jaipur', 'India', 'JAI'),
  ('c0000001-0000-0000-0000-000000000011', 'Lucknow', 'India', 'LKO'),
  ('c0000001-0000-0000-0000-000000000012', 'Kochi', 'India', 'COK')
ON CONFLICT (id) DO NOTHING;

-- Seed airports
INSERT INTO airports (id, name, city_id, code) VALUES
  ('a0000001-0000-0000-0000-000000000001', 'Chhatrapati Shivaji Maharaj International Airport', 'c0000001-0000-0000-0000-000000000001', 'BOM'),
  ('a0000001-0000-0000-0000-000000000002', 'Indira Gandhi International Airport', 'c0000001-0000-0000-0000-000000000002', 'DEL'),
  ('a0000001-0000-0000-0000-000000000003', 'Kempegowda International Airport', 'c0000001-0000-0000-0000-000000000003', 'BLR'),
  ('a0000001-0000-0000-0000-000000000004', 'Chennai International Airport', 'c0000001-0000-0000-0000-000000000004', 'MAA'),
  ('a0000001-0000-0000-0000-000000000005', 'Netaji Subhas Chandra Bose International Airport', 'c0000001-0000-0000-0000-000000000005', 'CCU'),
  ('a0000001-0000-0000-0000-000000000006', 'Rajiv Gandhi International Airport', 'c0000001-0000-0000-0000-000000000006', 'HYD'),
  ('a0000001-0000-0000-0000-000000000007', 'Pune Airport', 'c0000001-0000-0000-0000-000000000007', 'PNQ'),
  ('a0000001-0000-0000-0000-000000000008', 'Sardar Vallabhbhai Patel International Airport', 'c0000001-0000-0000-0000-000000000008', 'AMD'),
  ('a0000001-0000-0000-0000-000000000009', 'Goa International Airport', 'c0000001-0000-0000-0000-000000000009', 'GOI'),
  ('a0000001-0000-0000-0000-000000000010', 'Jaipur International Airport', 'c0000001-0000-0000-0000-000000000010', 'JAI'),
  ('a0000001-0000-0000-0000-000000000011', 'Chaudhary Charan Singh International Airport', 'c0000001-0000-0000-0000-000000000011', 'LKO'),
  ('a0000001-0000-0000-0000-000000000012', 'Cochin International Airport', 'c0000001-0000-0000-0000-000000000012', 'COK')
ON CONFLICT (id) DO NOTHING;

-- Generate flights for all airport pairs for 365 days with 2-3 flights per route per day
-- Using a DO block for procedural generation
DO $$
DECLARE
  orig_rec RECORD;
  dest_rec RECORD;
  d DATE;
  base_price INT;
  carrier TEXT;
  carriers TEXT[] := ARRAY['6E','AI','SG','UK','QP','I5','G8'];
  flight_num TEXT;
  dep_hour INT;
  dep_min INT;
  dur_hours NUMERIC;
  dep_ts TIMESTAMPTZ;
  arr_ts TIMESTAMPTZ;
  price_cents BIGINT;
  slot INT;
  total_seats INT;
BEGIN
  FOR orig_rec IN SELECT id, code FROM airports LOOP
    FOR dest_rec IN SELECT id, code FROM airports WHERE id != orig_rec.id LOOP

      -- base price varies by route (deterministic from codes)
      base_price := 3000 + (ASCII(SUBSTRING(orig_rec.code,1,1)) * 37 + ASCII(SUBSTRING(dest_rec.code,1,1)) * 53) % 8000;

      FOR d IN SELECT generate_series(CURRENT_DATE, CURRENT_DATE + INTERVAL '364 days', '1 day')::date LOOP
        -- 3 flights per route per day: morning, afternoon, evening
        FOR slot IN 1..3 LOOP
          carrier := carriers[1 + (ASCII(SUBSTRING(orig_rec.code,1,1)) + slot) % array_length(carriers,1)];
          flight_num := carrier || '-' || (1000 + (ASCII(SUBSTRING(orig_rec.code,2,1)) * 7 + ASCII(SUBSTRING(dest_rec.code,2,1)) * 11 + slot * 100) % 9000);

          CASE slot
            WHEN 1 THEN dep_hour := 6 + (ASCII(SUBSTRING(orig_rec.code,3,1))) % 3; dep_min := (ASCII(SUBSTRING(dest_rec.code,2,1)) * 7) % 60;
            WHEN 2 THEN dep_hour := 12 + (ASCII(SUBSTRING(orig_rec.code,2,1))) % 3; dep_min := (ASCII(SUBSTRING(dest_rec.code,3,1)) * 11) % 60;
            WHEN 3 THEN dep_hour := 18 + (ASCII(SUBSTRING(dest_rec.code,1,1))) % 3; dep_min := (ASCII(SUBSTRING(orig_rec.code,1,1)) * 13) % 60;
          END CASE;

          dur_hours := 1.5 + (ASCII(SUBSTRING(orig_rec.code,1,1)) + ASCII(SUBSTRING(dest_rec.code,1,1))) % 3 + (RANDOM() * 0.5);
          dep_ts := (d + make_interval(hours => dep_hour, mins => dep_min))::timestamptz;
          arr_ts := dep_ts + make_interval(mins => (dur_hours * 60)::int);

          -- price varies by slot and day of week
          price_cents := (base_price + slot * 400 + EXTRACT(DOW FROM d)::int * 200 + (RANDOM() * 1500)::int) * 100;
          total_seats := 150 + (slot * 30);

          INSERT INTO flights (flight_number, origin_id, destination_id, departure_time, arrival_time, price, seats_total, seats_available)
          VALUES (flight_num, orig_rec.id, dest_rec.id, dep_ts, arr_ts, price_cents, total_seats, total_seats)
          ON CONFLICT DO NOTHING;
        END LOOP;
      END LOOP;

    END LOOP;
  END LOOP;
END $$;
