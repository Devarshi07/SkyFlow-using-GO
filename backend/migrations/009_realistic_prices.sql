-- Update flight prices to realistic ranges (stored in cents)
-- US domestic: $79–$399
-- India domestic: $49–$299

UPDATE flights f
SET price = sub.new_price
FROM (
  SELECT
    f2.id,
    GREATEST(4900, LEAST(50000,
      CASE
        -- US domestic routes (JFK, LAX, ORD, SFO, MIA, DFW, SEA, BOS, DEN, LAS)
        WHEN o.code IN ('JFK','LAX','ORD','SFO','MIA','DFW','SEA','BOS','DEN','LAS')
         AND d.code IN ('JFK','LAX','ORD','SFO','MIA','DFW','SEA','BOS','DEN','LAS')
        THEN (
          7900
          + (ASCII(SUBSTRING(o.code, 1, 1)) * 37 + ASCII(SUBSTRING(d.code, 1, 1)) * 53) % 25000
          + EXTRACT(HOUR FROM f2.departure_time)::int * 50
          + EXTRACT(DOW FROM f2.departure_time)::int * 300
        )
        -- India domestic routes
        ELSE (
          4900
          + (ASCII(SUBSTRING(o.code, 1, 1)) * 31 + ASCII(SUBSTRING(d.code, 1, 1)) * 47) % 20000
          + EXTRACT(HOUR FROM f2.departure_time)::int * 40
          + EXTRACT(DOW FROM f2.departure_time)::int * 200
        )
      END
    )) AS new_price
  FROM flights f2
  JOIN airports o ON f2.origin_id = o.id
  JOIN airports d ON f2.destination_id = d.id
) sub
WHERE f.id = sub.id;
