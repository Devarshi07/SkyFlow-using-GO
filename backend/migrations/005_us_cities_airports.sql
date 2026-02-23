-- US Cities and Airports for domestic US flight search
INSERT INTO cities (id, name, country, code) VALUES
  ('c0000002-0000-0000-0000-000000000001', 'New York', 'United States', 'JFK'),
  ('c0000002-0000-0000-0000-000000000002', 'Los Angeles', 'United States', 'LAX'),
  ('c0000002-0000-0000-0000-000000000003', 'Chicago', 'United States', 'ORD'),
  ('c0000002-0000-0000-0000-000000000004', 'San Francisco', 'United States', 'SFO'),
  ('c0000002-0000-0000-0000-000000000005', 'Miami', 'United States', 'MIA'),
  ('c0000002-0000-0000-0000-000000000006', 'Dallas', 'United States', 'DFW'),
  ('c0000002-0000-0000-0000-000000000007', 'Seattle', 'United States', 'SEA'),
  ('c0000002-0000-0000-0000-000000000008', 'Boston', 'United States', 'BOS'),
  ('c0000002-0000-0000-0000-000000000009', 'Denver', 'United States', 'DEN'),
  ('c0000002-0000-0000-0000-000000000010', 'Las Vegas', 'United States', 'LAS')
ON CONFLICT (id) DO NOTHING;

INSERT INTO airports (id, name, city_id, code) VALUES
  ('a0000002-0000-0000-0000-000000000001', 'John F. Kennedy International Airport', 'c0000002-0000-0000-0000-000000000001', 'JFK'),
  ('a0000002-0000-0000-0000-000000000002', 'Los Angeles International Airport', 'c0000002-0000-0000-0000-000000000002', 'LAX'),
  ('a0000002-0000-0000-0000-000000000003', 'O''Hare International Airport', 'c0000002-0000-0000-0000-000000000003', 'ORD'),
  ('a0000002-0000-0000-0000-000000000004', 'San Francisco International Airport', 'c0000002-0000-0000-0000-000000000004', 'SFO'),
  ('a0000002-0000-0000-0000-000000000005', 'Miami International Airport', 'c0000002-0000-0000-0000-000000000005', 'MIA'),
  ('a0000002-0000-0000-0000-000000000006', 'Dallas/Fort Worth International Airport', 'c0000002-0000-0000-0000-000000000006', 'DFW'),
  ('a0000002-0000-0000-0000-000000000007', 'Seattle-Tacoma International Airport', 'c0000002-0000-0000-0000-000000000007', 'SEA'),
  ('a0000002-0000-0000-0000-000000000008', 'Boston Logan International Airport', 'c0000002-0000-0000-0000-000000000008', 'BOS'),
  ('a0000002-0000-0000-0000-000000000009', 'Denver International Airport', 'c0000002-0000-0000-0000-000000000009', 'DEN'),
  ('a0000002-0000-0000-0000-000000000010', 'Harry Reid International Airport', 'c0000002-0000-0000-0000-000000000010', 'LAS')
ON CONFLICT (id) DO NOTHING;

-- Generate flights for US airport pairs
DO $$
DECLARE
  orig_rec RECORD;
  dest_rec RECORD;
  d DATE;
  base_price INT;
  carrier TEXT;
  carriers TEXT[] := ARRAY['AA','UA','DL','WN','B6','AS','NK'];
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
  FOR orig_rec IN SELECT id, code FROM airports WHERE code IN ('JFK','LAX','ORD','SFO','MIA','DFW','SEA','BOS','DEN','LAS') LOOP
    FOR dest_rec IN SELECT id, code FROM airports WHERE code IN ('JFK','LAX','ORD','SFO','MIA','DFW','SEA','BOS','DEN','LAS') AND id != orig_rec.id LOOP

      base_price := 8000 + (ASCII(SUBSTRING(orig_rec.code,1,1)) * 37 + ASCII(SUBSTRING(dest_rec.code,1,1)) * 53) % 15000;

      FOR d IN SELECT generate_series(CURRENT_DATE, CURRENT_DATE + INTERVAL '364 days', '1 day')::date LOOP
        FOR slot IN 1..3 LOOP
          carrier := carriers[1 + (ASCII(SUBSTRING(orig_rec.code,1,1)) + slot) % array_length(carriers,1)];
          flight_num := carrier || (100 + (ASCII(SUBSTRING(orig_rec.code,2,1)) * 7 + ASCII(SUBSTRING(dest_rec.code,2,1)) * 11 + slot * 10) % 900);

          CASE slot
            WHEN 1 THEN dep_hour := 6 + (ASCII(SUBSTRING(orig_rec.code,3,1))) % 3; dep_min := (ASCII(SUBSTRING(dest_rec.code,2,1)) * 7) % 60;
            WHEN 2 THEN dep_hour := 12 + (ASCII(SUBSTRING(orig_rec.code,2,1))) % 3; dep_min := (ASCII(SUBSTRING(dest_rec.code,3,1)) * 11) % 60;
            WHEN 3 THEN dep_hour := 18 + (ASCII(SUBSTRING(dest_rec.code,1,1))) % 3; dep_min := (ASCII(SUBSTRING(orig_rec.code,1,1)) * 13) % 60;
          END CASE;

          dur_hours := 2.0 + (ASCII(SUBSTRING(orig_rec.code,1,1)) + ASCII(SUBSTRING(dest_rec.code,1,1))) % 4 + (RANDOM() * 1.0);
          dep_ts := (d + make_interval(hours => dep_hour, mins => dep_min))::timestamptz;
          arr_ts := dep_ts + make_interval(mins => (dur_hours * 60)::int);

          price_cents := (base_price + slot * 500 + EXTRACT(DOW FROM d)::int * 300 + (RANDOM() * 2000)::int) * 100;
          total_seats := 150 + (slot * 30);

          INSERT INTO flights (flight_number, origin_id, destination_id, departure_time, arrival_time, price, seats_total, seats_available)
          VALUES (flight_num, orig_rec.id, dest_rec.id, dep_ts, arr_ts, price_cents, total_seats, total_seats);
        END LOOP;
      END LOOP;

    END LOOP;
  END LOOP;
END $$;
