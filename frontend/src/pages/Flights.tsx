import { useState, useEffect, useMemo } from 'react';
import { useSearchParams, useNavigate } from 'react-router-dom';
import { gqlApi } from '../api/graphql';
import { bookingsApi, profileApi, type Flight, type ConnectingFlight, type Airport, type City, ApiError } from '../api/client';
import { useAuth } from '../context/AuthContext';
import './Flights.css';

type SortMode = 'price-asc' | 'price-desc' | 'depart-asc' | 'duration-asc';
type StopFilter = 'all' | 'nonstop' | '1stop';

interface UnifiedItem {
  type: 'direct' | 'connecting';
  flight?: Flight;
  connecting?: ConnectingFlight;
  price: number;
  departureTime: string;
  durationMin: number;
}

function durationMinutes(dep: string, arr: string) {
  return (new Date(arr).getTime() - new Date(dep).getTime()) / 60000;
}

function formatTime(iso: string) {
  return new Date(iso).toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit', hour12: true });
}

function formatDuration(min: number) {
  const h = Math.floor(min / 60);
  const m = Math.round(min % 60);
  return `${h}h ${m}m`;
}

function airportLabel(id: string, airports: Airport[], cities: City[]) {
  const ap = airports.find(a => a.id === id);
  if (!ap) return id.slice(0, 8);
  const city = cities.find(c => c.id === ap.city_id);
  return city ? `${ap.code} (${city.name})` : ap.code;
}

export function Flights() {
  const [searchParams, setSearchParams] = useSearchParams();
  const navigate = useNavigate();
  const { isLoggedIn } = useAuth();

  const [airports, setAirports] = useState<Airport[]>([]);
  const [cities, setCities] = useState<City[]>([]);
  const [origin, setOrigin] = useState(searchParams.get('origin') || '');
  const [dest, setDest] = useState(searchParams.get('destination') || '');
  const [date, setDate] = useState(searchParams.get('date') || new Date().toISOString().split('T')[0]);
  const [directFlights, setDirectFlights] = useState<Flight[]>([]);
  const [connectingFlights, setConnectingFlights] = useState<ConnectingFlight[]>([]);
  const [loading, setLoading] = useState(false);
  const [searched, setSearched] = useState(false);

  const [sortMode, setSortMode] = useState<SortMode>('price-asc');
  const [stopFilter, setStopFilter] = useState<StopFilter>('all');

  const [bookingFlight, setBookingFlight] = useState<Flight | null>(null);
  const [bookingPriceOverride, setBookingPriceOverride] = useState<number | null>(null);

  useEffect(() => {
    gqlApi.airportsAndCities().then(d => {
      setAirports(d.airports || []);
      setCities(d.cities || []);
    }).catch(() => {});
  }, []);

  useEffect(() => {
    const o = searchParams.get('origin');
    const d = searchParams.get('destination');
    const dt = searchParams.get('date');
    if (o && d && dt) {
      setOrigin(o);
      setDest(d);
      setDate(dt);
      doSearch(o, d, dt);
    }
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [searchParams.toString()]);

  async function doSearch(o: string, d: string, dt: string) {
    setLoading(true);
    setSearched(true);
    try {
      const res = await gqlApi.searchFlights(o, d, dt);
      setDirectFlights(res.searchFlights.flights || []);
      setConnectingFlights(res.searchFlights.connecting || []);
    } catch {
      setDirectFlights([]);
      setConnectingFlights([]);
    } finally {
      setLoading(false);
    }
  }

  function handleSearch(e: React.FormEvent) {
    e.preventDefault();
    if (!origin || !dest || !date) return;
    setSearchParams({ origin, destination: dest, date });
  }

  function shiftDate(days: number) {
    const d = new Date(date);
    d.setDate(d.getDate() + days);
    const newDate = d.toISOString().split('T')[0];
    setDate(newDate);
    if (origin && dest) {
      setSearchParams({ origin, destination: dest, date: newDate });
    }
  }

  const unifiedList = useMemo<UnifiedItem[]>(() => {
    const items: UnifiedItem[] = [];
    for (const f of directFlights) {
      items.push({
        type: 'direct',
        flight: f,
        price: f.price,
        departureTime: f.departure_time,
        durationMin: durationMinutes(f.departure_time, f.arrival_time),
      });
    }
    for (const c of connectingFlights) {
      items.push({
        type: 'connecting',
        connecting: c,
        price: c.total_price,
        departureTime: c.leg1.departure_time,
        durationMin: c.total_duration_hours * 60,
      });
    }

    let filtered = items;
    if (stopFilter === 'nonstop') filtered = items.filter(i => i.type === 'direct');
    if (stopFilter === '1stop') filtered = items.filter(i => i.type === 'connecting');

    filtered.sort((a, b) => {
      switch (sortMode) {
        case 'price-asc': return a.price - b.price;
        case 'price-desc': return b.price - a.price;
        case 'depart-asc': return new Date(a.departureTime).getTime() - new Date(b.departureTime).getTime();
        case 'duration-asc': return a.durationMin - b.durationMin;
        default: return 0;
      }
    });
    return filtered;
  }, [directFlights, connectingFlights, sortMode, stopFilter]);

  return (
    <div className="flights-page">
      <form className="flights-search-bar" onSubmit={handleSearch}>
        <select value={origin} onChange={e => setOrigin(e.target.value)} required>
          <option value="">From</option>
          {airports.map(a => {
            const city = cities.find(c => c.id === a.city_id);
            return <option key={a.id} value={a.id}>{a.code} - {city?.name || a.name}</option>;
          })}
        </select>
        <select value={dest} onChange={e => setDest(e.target.value)} required>
          <option value="">To</option>
          {airports.filter(a => a.id !== origin).map(a => {
            const city = cities.find(c => c.id === a.city_id);
            return <option key={a.id} value={a.id}>{a.code} - {city?.name || a.name}</option>;
          })}
        </select>
        <div className="date-nav">
          <button type="button" className="date-arrow" onClick={() => shiftDate(-1)}>◀</button>
          <input type="date" value={date} onChange={e => setDate(e.target.value)} required />
          <button type="button" className="date-arrow" onClick={() => shiftDate(1)}>▶</button>
        </div>
        <button type="submit" className="btn btn-primary">Search</button>
      </form>

      {searched && (
        <>
          <div className="sort-filter-bar">
            <div className="stop-chips">
              {(['all', 'nonstop', '1stop'] as StopFilter[]).map(f => (
                <button key={f} className={`stop-chip ${stopFilter === f ? 'active' : ''}`} onClick={() => setStopFilter(f)}>
                  {f === 'all' ? 'All' : f === 'nonstop' ? 'Non Stop' : '1 Stop'}
                </button>
              ))}
            </div>
            <select className="sort-select" value={sortMode} onChange={e => setSortMode(e.target.value as SortMode)}>
              <option value="price-asc">Price: Low to High</option>
              <option value="price-desc">Price: High to Low</option>
              <option value="depart-asc">Departure: Earliest</option>
              <option value="duration-asc">Duration: Shortest</option>
            </select>
          </div>

          {loading ? (
            <div className="flights-loading"><span className="spinner" /> Searching flights...</div>
          ) : unifiedList.length === 0 ? (
            <div className="no-flights">No flights found for this route and date.</div>
          ) : (
            <>
              <div className="results-count">{unifiedList.length} flight{unifiedList.length !== 1 ? 's' : ''} found</div>
              {unifiedList.map((item, i) =>
                item.type === 'direct' && item.flight ? (
                  <DirectFlightCard
                    key={`d-${item.flight.id}`}
                    flight={item.flight}
                    airports={airports}
                    cities={cities}
                    onBook={() => {
                      if (!isLoggedIn) { navigate('/login'); return; }
                      setBookingFlight(item.flight!);
                      setBookingPriceOverride(null);
                    }}
                  />
                ) : item.connecting ? (
                  <ConnectingFlightCard
                    key={`c-${i}`}
                    cf={item.connecting}
                    airports={airports}
                    cities={cities}
                    onBook={() => {
                      if (!isLoggedIn) { navigate('/login'); return; }
                      setBookingFlight(item.connecting!.leg1);
                      setBookingPriceOverride(item.connecting!.total_price);
                    }}
                  />
                ) : null
              )}
            </>
          )}
        </>
      )}

      {bookingFlight && (
        <BookFlightModal
          flight={bookingFlight}
          displayPrice={bookingPriceOverride ?? bookingFlight.price}
          onClose={() => { setBookingFlight(null); setBookingPriceOverride(null); }}
        />
      )}
    </div>
  );
}

function DirectFlightCard({ flight, airports, cities, onBook }: {
  flight: Flight;
  airports: Airport[];
  cities: City[];
  onBook: () => void;
}) {
  const dur = durationMinutes(flight.departure_time, flight.arrival_time);
  return (
    <div className="flight-card">
      <div className="flight-card-main">
        <div className="flight-info-col">
          <span className="flight-number">{flight.flight_number}</span>
          <span className="stop-badge nonstop">Non Stop</span>
        </div>
        <div className="flight-route-col">
          <div className="flight-time-group">
            <span className="flight-time">{formatTime(flight.departure_time)}</span>
            <span className="flight-airport">{airportLabel(flight.origin_id, airports, cities)}</span>
          </div>
          <div className="flight-duration-line">
            <span className="flight-duration">{formatDuration(dur)}</span>
            <div className="duration-line" />
          </div>
          <div className="flight-time-group">
            <span className="flight-time">{formatTime(flight.arrival_time)}</span>
            <span className="flight-airport">{airportLabel(flight.destination_id, airports, cities)}</span>
          </div>
        </div>
        <div className="flight-price-col">
          <span className="flight-price">${(flight.price / 100).toFixed(2)}</span>
          <span className="flight-seats">{flight.seats_available} seats left</span>
          <button className="btn btn-primary book-btn" onClick={onBook}>Book</button>
        </div>
      </div>
    </div>
  );
}

function ConnectingFlightCard({ cf, airports, cities, onBook }: {
  cf: ConnectingFlight;
  airports: Airport[];
  cities: City[];
  onBook: () => void;
}) {
  return (
    <div className="flight-card connecting-card">
      <div className="flight-card-main">
        <div className="flight-info-col">
          <span className="flight-number">{cf.leg1.flight_number}</span>
          <span className="stop-badge onestop">1 Stop</span>

        </div>
        <div className="flight-route-col">
          <div className="flight-time-group">
            <span className="flight-time">{formatTime(cf.leg1.departure_time)}</span>
            <span className="flight-airport">{airportLabel(cf.leg1.origin_id, airports, cities)}</span>
          </div>
          <div className="flight-duration-line">
            <span className="flight-duration">{cf.total_duration_hours.toFixed(1)}h total</span>
            <div className="duration-line" />
            <span className="layover-info">{formatDuration(cf.layover_minutes)} layover at {airportLabel(cf.layover_airport_id, airports, cities)}</span>
          </div>
          <div className="flight-time-group">
            <span className="flight-time">{formatTime(cf.leg2.arrival_time)}</span>
            <span className="flight-airport">{airportLabel(cf.leg2.destination_id, airports, cities)}</span>
          </div>
        </div>
        <div className="flight-price-col">
          <span className="flight-price">${(cf.total_price / 100).toFixed(2)}</span>
          <button className="btn btn-primary book-btn" onClick={onBook}>Book</button>
        </div>
      </div>
    </div>
  );
}

function BookFlightModal({ flight, displayPrice, onClose }: { flight: Flight; displayPrice: number; onClose: () => void }) {
  const navigate = useNavigate();
  const { user } = useAuth();
  const [name, setName] = useState('');
  const [email, setEmail] = useState('');
  const [phone, setPhone] = useState('');
  const [seats, setSeats] = useState(1);
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (user) {
      setEmail(user.email || '');
      if (user.full_name) setName(user.full_name);
      profileApi.get().then(p => {
        if (p.full_name) setName(p.full_name);
        if (p.phone) setPhone(p.phone);
        if (p.email) setEmail(p.email);
      }).catch(() => {});
    }
  }, [user]);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError('');
    setLoading(true);
    try {
      const res = await bookingsApi.create({
        flight_id: flight.id,
        seats,
        passenger_name: name,
        passenger_email: email,
        passenger_phone: phone,
      });
      // Stripe live mode — redirect to Stripe Checkout
      if (res.checkout_url) {
        window.location.href = res.checkout_url;
        return;
      }
      // Demo mode — use built-in checkout page
      // Only show demo checkout if payment_intent_id starts with "pi_mock"
      if (res.payment_intent_id && res.payment_intent_id.startsWith('pi_mock')) {
        const params = new URLSearchParams();
        params.set('booking_id', res.booking_id);
        params.set('payment_intent_id', res.payment_intent_id);
        params.set('amount', res.amount.toString());
        params.set('flight_id', flight.id);
        if (name) params.set('name', name);
        if (email) params.set('email', email);
        navigate(`/checkout?${params.toString()}`);
        return;
      }
      // If we have a real PI but no checkout_url, something went wrong
      if (res.payment_intent_id) {
        setError('Payment setup failed — please try again');
        setLoading(false);
        return;
      }
      navigate(`/bookings/${res.booking_id}`);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Booking failed');
      setLoading(false);
    }
  }

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal-content" onClick={e => e.stopPropagation()}>
        <h2>Book Flight {flight.flight_number}</h2>
        <p style={{ color: 'var(--text-muted)', marginBottom: '1rem' }}>
          ${(displayPrice / 100).toFixed(2)} per seat
        </p>
        {error && <div className="auth-error">{error}</div>}
        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <label>Passenger Name</label>
            <input value={name} onChange={e => setName(e.target.value)} required />
          </div>
          <div className="form-group">
            <label>Email</label>
            <input type="email" value={email} onChange={e => setEmail(e.target.value)} required />
          </div>
          <div className="form-group">
            <label>Phone</label>
            <input type="tel" value={phone} onChange={e => setPhone(e.target.value)} />
          </div>
          <div className="form-group">
            <label>Seats</label>
            <input type="number" min={1} max={flight.seats_available} value={seats} onChange={e => setSeats(Number(e.target.value))} />
          </div>
          <div style={{ display: 'flex', gap: '0.75rem', marginTop: '1rem' }}>
            <button type="button" className="btn btn-secondary" onClick={onClose} style={{ flex: 1 }}>Cancel</button>
            <button type="submit" className="btn btn-primary" disabled={loading} style={{ flex: 1 }}>
              {loading ? <span className="spinner" /> : 'Continue to Payment'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
