import { useState, useEffect, useMemo } from 'react';
import { useSearchParams, useNavigate } from 'react-router-dom';
import { gqlApi } from '../api/graphql';
import { flightsApi, profileApi, type Flight, type ConnectingFlight, type Airport, type City } from '../api/client';
import { SkyFlowLogo } from '../components/SkyFlowLogo';
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

  // Round-trip state
  const isRoundTrip = searchParams.get('trip') === 'round';
  const returnDate = searchParams.get('return_date') || '';
  const [roundTripStep, setRoundTripStep] = useState<'outbound' | 'return'>('outbound');
  const [selectedOutbound, setSelectedOutbound] = useState<Flight | null>(null);
  const [selectedOutboundPrice, setSelectedOutboundPrice] = useState<number | null>(null);
  // Return flight results
  const [returnDirectFlights, setReturnDirectFlights] = useState<Flight[]>([]);
  const [returnConnectingFlights, setReturnConnectingFlights] = useState<ConnectingFlight[]>([]);
  const [returnLoading, setReturnLoading] = useState(false);

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

  // Auto-open booking modal if returning from login with a `book` param
  useEffect(() => {
    if (!isLoggedIn) return;
    const bookFlightId = searchParams.get('book');
    if (!bookFlightId) return;
    const priceOverride = searchParams.get('price');

    flightsApi.get(bookFlightId).then(f => {
      setBookingFlight(f);
      setBookingPriceOverride(priceOverride ? parseInt(priceOverride, 10) : null);
      // Clean the book param from URL
      const newParams = new URLSearchParams(searchParams);
      newParams.delete('book');
      newParams.delete('price');
      setSearchParams(newParams, { replace: true });
    }).catch(() => {});
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isLoggedIn, searchParams.get('book')]);

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
    setRoundTripStep('outbound');
    setSelectedOutbound(null);
    setSelectedOutboundPrice(null);
    const params: Record<string, string> = { origin, destination: dest, date };
    if (isRoundTrip && returnDate) {
      params.trip = 'round';
      params.return_date = returnDate;
    }
    setSearchParams(params);
  }

  function handleSelectOutbound(flight: Flight, priceOverride?: number) {
    setSelectedOutbound(flight);
    setSelectedOutboundPrice(priceOverride ?? null);
    setRoundTripStep('return');
    // Search return flights
    setReturnLoading(true);
    gqlApi.searchFlights(dest, origin, returnDate).then(res => {
      setReturnDirectFlights(res.searchFlights.flights || []);
      setReturnConnectingFlights(res.searchFlights.connecting || []);
    }).catch(() => {
      setReturnDirectFlights([]);
      setReturnConnectingFlights([]);
    }).finally(() => setReturnLoading(false));
  }

  function handleSelectReturn(flight: Flight, priceOverride?: number) {
    // Open booking modal with both flights
    setBookingFlight(flight);
    setBookingPriceOverride(priceOverride ?? null);
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

  const returnUnifiedList = useMemo<UnifiedItem[]>(() => {
    const items: UnifiedItem[] = [];
    for (const f of returnDirectFlights) {
      items.push({ type: 'direct', flight: f, price: f.price, departureTime: f.departure_time, durationMin: durationMinutes(f.departure_time, f.arrival_time) });
    }
    for (const c of returnConnectingFlights) {
      items.push({ type: 'connecting', connecting: c, price: c.total_price, departureTime: c.leg1.departure_time, durationMin: c.total_duration_hours * 60 });
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
  }, [returnDirectFlights, returnConnectingFlights, sortMode, stopFilter]);

  // Determine which list and handlers to show
  const showingReturn = isRoundTrip && roundTripStep === 'return';
  const activeList = showingReturn ? returnUnifiedList : unifiedList;
  const activeLoading = showingReturn ? returnLoading : loading;

  function getCityName(airportId: string) {
    const ap = airports.find(a => a.id === airportId);
    if (!ap) return '';
    return cities.find(c => c.id === ap.city_id)?.name || ap.code;
  }

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

      {/* Round-trip step indicator */}
      {isRoundTrip && searched && (
        <div className="rt-steps">
          <div className={`rt-step ${roundTripStep === 'outbound' ? 'active' : 'done'}`}>
            <span className="rt-step-num">{roundTripStep === 'return' ? '✓' : '1'}</span>
            <div>
              <div className="rt-step-title">Outbound · {new Date(date).toLocaleDateString('en-US', { month: 'short', day: 'numeric' })}</div>
              <div className="rt-step-sub">{getCityName(origin)} → {getCityName(dest)}</div>
            </div>
            {selectedOutbound && roundTripStep === 'return' && (
              <div className="rt-step-selected">
                {formatTime(selectedOutbound.departure_time)} · {selectedOutbound.flight_number} · ${((selectedOutboundPrice ?? selectedOutbound.price) / 100).toFixed(0)}
                <button className="rt-change-btn" onClick={() => { setRoundTripStep('outbound'); setSelectedOutbound(null); }}>Change</button>
              </div>
            )}
          </div>
          <div className={`rt-step ${roundTripStep === 'return' ? 'active' : ''}`}>
            <span className="rt-step-num">2</span>
            <div>
              <div className="rt-step-title">Return · {returnDate ? new Date(returnDate).toLocaleDateString('en-US', { month: 'short', day: 'numeric' }) : ''}</div>
              <div className="rt-step-sub">{getCityName(dest)} → {getCityName(origin)}</div>
            </div>
          </div>
        </div>
      )}

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

          {activeLoading ? (
            <div className="flights-loading"><span className="spinner" /> {showingReturn ? 'Searching return flights...' : 'Searching flights...'}</div>
          ) : activeList.length === 0 ? (
            <div className="no-flights">No flights found for this route and date.</div>
          ) : (
            <>
              <div className="results-count">{activeList.length} flight{activeList.length !== 1 ? 's' : ''} found</div>
              {activeList.map((item, i) => {
                function handleBook(flight: Flight, priceOvr?: number) {
                  if (!isLoggedIn) {
                    const returnUrl = `/flights?origin=${origin}&destination=${dest}&date=${date}&book=${flight.id}${priceOvr ? `&price=${priceOvr}` : ''}${isRoundTrip ? `&trip=round&return_date=${returnDate}` : ''}`;
                    navigate(`/login?redirect=${encodeURIComponent(returnUrl)}`);
                    return;
                  }
                  if (isRoundTrip && roundTripStep === 'outbound') {
                    handleSelectOutbound(flight, priceOvr);
                  } else if (isRoundTrip && roundTripStep === 'return') {
                    handleSelectReturn(flight, priceOvr);
                  } else {
                    setBookingFlight(flight);
                    setBookingPriceOverride(priceOvr ?? null);
                  }
                }
                return item.type === 'direct' && item.flight ? (
                  <DirectFlightCard
                    key={`d-${item.flight.id}`}
                    flight={item.flight}
                    airports={airports}
                    cities={cities}
                    onBook={() => handleBook(item.flight!)}
                    btnLabel={isRoundTrip ? 'Select' : 'Book'}
                  />
                ) : item.connecting ? (
                  <ConnectingFlightCard
                    key={`c-${i}`}
                    cf={item.connecting}
                    airports={airports}
                    cities={cities}
                    onBook={() => handleBook(item.connecting!.leg1, item.connecting!.total_price)}
                    btnLabel={isRoundTrip ? 'Select' : 'Book'}
                  />
                ) : null;
              })}
            </>
          )}
        </>
      )}

      {bookingFlight && (
        <BookFlightModal
          flight={bookingFlight}
          displayPrice={bookingPriceOverride ?? bookingFlight.price}
          outboundFlight={isRoundTrip ? selectedOutbound : null}
          isRoundTrip={isRoundTrip}
          onClose={() => { setBookingFlight(null); setBookingPriceOverride(null); }}
        />
      )}
    </div>
  );
}

function DirectFlightCard({ flight, airports, cities, onBook, btnLabel = 'Book' }: {
  flight: Flight;
  airports: Airport[];
  cities: City[];
  onBook: () => void;
  btnLabel?: string;
}) {
  const dur = durationMinutes(flight.departure_time, flight.arrival_time);
  return (
    <div className="flight-card">
      <div className="flight-card-main">
        <div className="flight-info-col">
          <SkyFlowLogo size="sm" flightNumbers={flight.flight_number} />
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
          <button className="btn btn-primary book-btn" onClick={onBook}>{btnLabel}</button>
        </div>
      </div>
    </div>
  );
}

function ConnectingFlightCard({ cf, airports, cities, onBook, btnLabel = 'Book' }: {
  cf: ConnectingFlight;
  airports: Airport[];
  cities: City[];
  onBook: () => void;
  btnLabel?: string;
}) {
  return (
    <div className="flight-card connecting-card">
      <div className="flight-card-main">
        <div className="flight-info-col">
          <SkyFlowLogo size="sm" flightNumbers={`${cf.leg1.flight_number}, ${cf.leg2.flight_number}`} />
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
          <button className="btn btn-primary book-btn" onClick={onBook}>{btnLabel}</button>
        </div>
      </div>
    </div>
  );
}

function BookFlightModal({ flight, displayPrice, outboundFlight, isRoundTrip, onClose }: {
  flight: Flight;
  displayPrice: number;
  outboundFlight?: Flight | null;
  isRoundTrip?: boolean;
  onClose: () => void;
}) {
  const navigate = useNavigate();
  const { user } = useAuth();
  const [name, setName] = useState('');
  const [email, setEmail] = useState('');
  const [phone, setPhone] = useState('');
  const [seats, setSeats] = useState(1);
  const [error] = useState('');
  const [loading] = useState(false);

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

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!name || !email || !phone || phone.replace(/\D/g, '').length < 7) return;
    const params = new URLSearchParams();
    if (isRoundTrip && outboundFlight) {
      params.set('flight_id', outboundFlight.id);
      params.set('return_flight_id', flight.id);
      params.set('trip', 'round');
    } else {
      params.set('flight_id', flight.id);
    }
    params.set('seats', seats.toString());
    params.set('name', name);
    params.set('email', email);
    params.set('phone', phone);
    navigate(`/review?${params.toString()}`);
  }

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal-content" onClick={e => e.stopPropagation()}>
        <h2>{isRoundTrip ? 'Book Round Trip' : `Book Flight ${flight.flight_number}`}</h2>
        <p style={{ color: 'var(--text-muted)', marginBottom: '1rem' }}>
          {isRoundTrip && outboundFlight
            ? `Outbound: ${outboundFlight.flight_number} · Return: ${flight.flight_number}`
            : `${(displayPrice / 100).toFixed(2)} per seat`
          }
        </p>
        {error && <div className="auth-error">{error}</div>}
        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <label>Passenger Name <span style={{ color: 'var(--danger)', fontWeight: 400 }}>*</span></label>
            <input value={name} onChange={e => setName(e.target.value)} required />
          </div>
          <div className="form-group">
            <label>Email <span style={{ color: 'var(--danger)', fontWeight: 400 }}>*</span></label>
            <input type="email" value={email} onChange={e => setEmail(e.target.value)} required />
          </div>
          <div className="form-group">
            <label>Phone <span style={{ color: 'var(--danger)', fontWeight: 400 }}>*</span></label>
            <input type="tel" value={phone} onChange={e => setPhone(e.target.value.replace(/[^\d+\-() ]/g, '').slice(0, 20))} placeholder="+1 (555) 000-0000" required />
            {phone.length > 0 && phone.replace(/\D/g, '').length < 7 && (
              <div className="field-error">Enter a valid phone number</div>
            )}
          </div>
          <div className="form-group">
            <label>Seats <span style={{ color: 'var(--danger)', fontWeight: 400 }}>*</span></label>
            <input type="number" min={1} max={flight.seats_available} value={seats} onChange={e => setSeats(Number(e.target.value))} />
          </div>
          <div style={{ display: 'flex', gap: '0.75rem', marginTop: '1rem' }}>
            <button type="button" className="btn btn-secondary" onClick={onClose} style={{ flex: 1 }}>Cancel</button>
            <button type="submit" className="btn btn-primary" disabled={loading} style={{ flex: 1 }}>
              {loading ? <span className="spinner" /> : 'Review Booking →'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
