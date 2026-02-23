import { useState, useEffect, useMemo, useRef, forwardRef } from 'react';
import { useNavigate } from 'react-router-dom';
import DatePicker from 'react-datepicker';
import 'react-datepicker/dist/react-datepicker.css';
import { gqlApi } from '../api/graphql';
import type { Airport, City } from '../api/client';
import './Home.css';

const TRENDING_PAIRS = [
  ['JFK', 'LAX'], ['LAX', 'SFO'], ['ORD', 'LAX'], ['JFK', 'MIA'],
  ['SFO', 'SEA'], ['DFW', 'DEN'], ['BOS', 'JFK'], ['MIA', 'ORD'],
];

const today = new Date();
today.setHours(0, 0, 0, 0);

const DateFieldInput = forwardRef<
  HTMLInputElement,
  { value?: string; onClick?: () => void; placeholder?: string; onOpenCalendar?: () => void }
>(({ value, onClick, placeholder = 'Date', onOpenCalendar }, ref) => {
  const handleOpen = () => {
    onOpenCalendar?.();
    onClick?.();
  };
  return (
    <div
      className="date-field-trigger"
      onClick={(e) => {
        e.preventDefault();
        handleOpen();
      }}
    >
      <input
        ref={ref}
        type="text"
        readOnly
        value={value || ''}
        placeholder={placeholder}
        onClick={(e) => {
          e.preventDefault();
          handleOpen();
        }}
        className={`date-input ${!value ? 'placeholder' : ''}`}
        aria-label="Select date"
      />
      <span className="field-chevron" aria-hidden>▾</span>
    </div>
  );
});

DateFieldInput.displayName = 'DateFieldInput';

export function Home() {
  const navigate = useNavigate();
  const [airports, setAirports] = useState<Airport[]>([]);
  const [cities, setCities] = useState<City[]>([]);
  const [origin, setOrigin] = useState('');
  const [dest, setDest] = useState('');
  const [date, setDate] = useState('');
  const [tripType, setTripType] = useState<'oneway' | 'round'>('oneway');
  const [returnDate, setReturnDate] = useState('');
  const [travellers, setTravellers] = useState(1);
  const [cabinClass, setCabinClass] = useState<'Economy' | 'Premium Economy' | 'Business'>('Economy');
  const [travellerOpen, setTravellerOpen] = useState(false);
  const [departureOpen, setDepartureOpen] = useState(false);
  const [returnOpen, setReturnOpen] = useState(false);
  const travellerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    function handleClickOutside(e: MouseEvent) {
      if (travellerRef.current && !travellerRef.current.contains(e.target as Node)) {
        setTravellerOpen(false);
      }
    }
    if (travellerOpen) document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, [travellerOpen]);

  useEffect(() => {
    gqlApi.airportsAndCities().then(d => {
      setAirports(d.airports || []);
      setCities(d.cities || []);
    }).catch(() => {});
  }, []);

  function cityForAirport(a: Airport) {
    return cities.find(c => c.id === a.city_id);
  }

  function airportByCode(code: string) {
    return airports.find(a => a.code === code);
  }

  const originAirport = airports.find(a => a.id === origin);
  const destAirport = airports.find(a => a.id === dest);
  const originCity = originAirport ? cityForAirport(originAirport) : null;
  const destCity = destAirport ? cityForAirport(destAirport) : null;

  function handleSearch(e: React.FormEvent) {
    e.preventDefault();
    if (!origin || !dest || !date) return;
    navigate(`/flights?origin=${origin}&destination=${dest}&date=${date}`);
  }

  function swapCities() {
    const tmp = origin;
    setOrigin(dest);
    setDest(tmp);
  }

  const trendingRoutes = useMemo(() => {
    if (airports.length === 0) return [];
    return TRENDING_PAIRS.map(([from, to]) => {
      const a1 = airportByCode(from);
      const a2 = airportByCode(to);
      if (!a1 || !a2) return null;
      const c1 = cityForAirport(a1);
      const c2 = cityForAirport(a2);
      return { from: a1, to: a2, fromCity: c1, toCity: c2 };
    }).filter(Boolean).slice(0, 4);
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [airports, cities]);

  const popularRoutes = useMemo(() => {
    if (airports.length === 0) return [];
    const pairs = [
      ['JFK', 'LAX', 28900], ['LAX', 'SFO', 9900], ['ORD', 'LAX', 19900],
      ['JFK', 'MIA', 15900], ['SFO', 'SEA', 12900], ['DFW', 'DEN', 11900],
    ];
    return pairs.map(([from, to, price]) => {
      const a1 = airportByCode(from as string);
      const a2 = airportByCode(to as string);
      if (!a1 || !a2) return null;
      const c1 = cityForAirport(a1);
      const c2 = cityForAirport(a2);
      return { from: a1, to: a2, fromCity: c1, toCity: c2, price: price as number };
    }).filter(Boolean);
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [airports, cities]);

  return (
    <div className="home-page">
      <section className="hero-section">
        <div className="hero-inner">
          <h1 className="hero-headline">Search & Book Flights</h1>
          <p className="hero-sub">Best prices on domestic and international flights across the US</p>

          <div className="search-card">
            <div className="search-card-header">
              <div className="trip-tabs">
                <label className={`trip-tab ${tripType === 'oneway' ? 'active' : ''}`}>
                  <input type="radio" name="trip" checked={tripType === 'oneway'} onChange={() => setTripType('oneway')} />
                  <span className="tab-dot" />
                  One Way
                </label>
                <label className={`trip-tab ${tripType === 'round' ? 'active' : ''}`}>
                  <input type="radio" name="trip" checked={tripType === 'round'} onChange={() => setTripType('round')} />
                  <span className="tab-dot" />
                  Round Trip
                </label>
              </div>
            </div>

            <form className="search-form" onSubmit={handleSearch}>
              <div className="search-fields">
                <div className="cities-row">
                  <div className="field-box from-field">
                    <label className="field-label">FROM</label>
                    <select value={origin} onChange={e => setOrigin(e.target.value)} className="field-select" required>
                      <option value="">Select city</option>
                      {airports.map(a => {
                        const c = cityForAirport(a);
                        return <option key={a.id} value={a.id}>{c?.name || a.name} ({a.code})</option>;
                      })}
                    </select>
                    {originAirport ? (
                      <div className="field-location">{originCity?.name || originAirport.name} ({originAirport.code})</div>
                    ) : (
                      <div className="field-location field-placeholder">Select city</div>
                    )}
                  </div>
                  <div className="swap-cell">
                    <button type="button" className="swap-btn" onClick={swapCities} title="Swap cities">
                      <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M17 8H3M14 5l3 3-3 3"/><path d="M7 16h14M10 19l-3-3 3-3"/></svg>
                    </button>
                  </div>
                  <div className="field-box to-field">
                  <label className="field-label">TO</label>
                  <select value={dest} onChange={e => setDest(e.target.value)} className="field-select" required>
                    <option value="">Select city</option>
                    {airports.filter(a => a.id !== origin).map(a => {
                      const c = cityForAirport(a);
                      return <option key={a.id} value={a.id}>{c?.name || a.name} ({a.code})</option>;
                    })}
                  </select>
                  {destAirport ? (
                    <div className="field-location">{destCity?.name || destAirport.name} ({destAirport.code})</div>
                  ) : (
                    <div className="field-location field-placeholder">Select city</div>
                  )}
                </div>
                </div>

                <div className="field-box date-field date-clickable">
                  <label className="field-label">DEPARTURE</label>
                  <DatePicker
                    selected={date ? new Date(date + 'T12:00:00') : null}
                    onChange={(d: Date | null) => {
                      setDate(d ? d.toISOString().split('T')[0] : '');
                      setDepartureOpen(false);
                    }}
                    minDate={today}
                    dateFormat="M/d/yy"
                    open={departureOpen}
                    onCalendarClose={() => setDepartureOpen(false)}
                    shouldCloseOnSelect
                    preventOpenOnFocus
                    customInput={<DateFieldInput placeholder="Date" onOpenCalendar={() => setDepartureOpen(true)} />}
                    showPopperArrow={false}
                    popperPlacement="bottom-start"
                    popperModifiers={[{ name: 'flip', options: { enabled: false } }] as never}
                  />
                </div>

                {tripType === 'round' && (
                  <div className="field-box date-field date-clickable">
                    <label className="field-label">RETURN</label>
                    <DatePicker
                      selected={returnDate ? new Date(returnDate + 'T12:00:00') : null}
                      onChange={(d: Date | null) => {
                        setReturnDate(d ? d.toISOString().split('T')[0] : '');
                        setReturnOpen(false);
                      }}
                      minDate={date ? new Date(date + 'T12:00:00') : today}
                      dateFormat="M/d/yy"
                      placeholderText="Date"
                      open={returnOpen}
                      onCalendarClose={() => setReturnOpen(false)}
                      shouldCloseOnSelect
                      preventOpenOnFocus
                      customInput={<DateFieldInput placeholder="Date" onOpenCalendar={() => setReturnOpen(true)} />}
                      showPopperArrow={false}
                      popperPlacement="bottom-start"
                      popperModifiers={[{ name: 'flip', options: { enabled: false } }] as never}
                    />
                  </div>
                )}

                <div className="field-box traveller-field" ref={travellerRef}>
                  <label className="field-label">TRAVELLERS & CLASS</label>
                  <div className="field-select-trigger" onClick={() => setTravellerOpen(!travellerOpen)}>
                    <div className="field-big">{travellers} <span className="traveller-label">{travellers === 1 ? 'Traveller' : 'Travellers'}</span></div>
                    <div className="field-detail">{cabinClass}</div>
                    <span className="field-chevron" aria-hidden>▾</span>
                  </div>
                  {travellerOpen && (
                    <div className="traveller-dropdown">
                      <div className="traveller-row">
                        <span>Travellers</span>
                        <div className="traveller-stepper">
                          <button type="button" onClick={() => setTravellers(Math.max(1, travellers - 1))}>−</button>
                          <span>{travellers}</span>
                          <button type="button" onClick={() => setTravellers(Math.min(9, travellers + 1))}>+</button>
                        </div>
                      </div>
                      <div className="traveller-row">
                        <span>Class</span>
                        <select value={cabinClass} onChange={e => setCabinClass(e.target.value as typeof cabinClass)} onClick={e => e.stopPropagation()}>
                          <option value="Economy">Economy</option>
                          <option value="Premium Economy">Premium Economy</option>
                          <option value="Business">Business</option>
                        </select>
                      </div>
                    </div>
                  )}
                </div>
              </div>

              <button type="submit" className="search-submit-btn">SEARCH</button>
            </form>
          </div>

          {trendingRoutes.length > 0 && (
            <div className="trending-bar">
              <span className="trending-label">Trending Searches:</span>
              {trendingRoutes.map((r: any, i: number) => (
                <button key={i} className="trending-pill" onClick={() => {
                  setOrigin(r.from.id);
                  setDest(r.to.id);
                  const d = date || new Date().toISOString().split('T')[0];
                  navigate(`/flights?origin=${r.from.id}&destination=${r.to.id}&date=${d}`);
                }}>
                  {r.fromCity?.name || r.from.code} &rarr; {r.toCity?.name || r.to.code}
                </button>
              ))}
            </div>
          )}
        </div>
      </section>

      <section className="promos-section container">
        <h2 className="promos-title">Offers & Promotions</h2>
        <div className="promos-grid">
          <div className="promo-card promo-card-accent">
            <div className="promo-card-icon">&#127915;</div>
            <h3>Up to 25% OFF</h3>
            <p>First booking discount. Sign in to unlock.</p>
            <span className="promo-cta">Sign in now &rarr;</span>
          </div>
          <div className="promo-card">
            <div className="promo-card-icon">&#128179;</div>
            <h3>PayPal Checkout</h3>
            <p>Pay with PayPal for fast, secure checkout. No account needed.</p>
          </div>
          <div className="promo-card">
            <div className="promo-card-icon">&#128179;</div>
            <h3>Visa & Mastercard</h3>
            <p>Earn 2x miles or points when you pay with eligible cards.</p>
          </div>
          <div className="promo-card">
            <div className="promo-card-icon">&#128179;</div>
            <h3>American Express Offers</h3>
            <p>Amex cardmembers get 15% back on select flight bookings.</p>
          </div>
          <div className="promo-card">
            <div className="promo-card-icon">&#128179;</div>
            <h3>Apple Pay & Google Pay</h3>
            <p>One-tap checkout. Instant payment, maximum security.</p>
          </div>
          <div className="promo-card">
            <div className="promo-card-icon">&#128274;</div>
            <h3>Secure Payments</h3>
            <p>Bank-level encryption with Stripe. Your data stays safe.</p>
          </div>
        </div>
      </section>

      <section className="popular-section container">
        <h2>Popular Routes</h2>
        <div className="popular-grid">
          {popularRoutes.map((r: any, i: number) => (
            <button key={i} className="popular-route-card" onClick={() => {
              setOrigin(r.from.id);
              setDest(r.to.id);
              const d = date || new Date().toISOString().split('T')[0];
              navigate(`/flights?origin=${r.from.id}&destination=${r.to.id}&date=${d}`);
            }}>
              <div className="route-names">
                <span className="route-city">{r.fromCity?.name}</span>
                <span className="route-arrow">&rarr;</span>
                <span className="route-city">{r.toCity?.name}</span>
              </div>
              <div className="route-codes">{r.from.code} &mdash; {r.to.code}</div>
              <div className="route-price">from <strong>${(r.price / 100).toFixed(0)}</strong></div>
            </button>
          ))}
        </div>
      </section>

      <section className="features-section container">
        <div className="features-grid">
          <div className="feature-item">
            <div className="feature-icon">&#128269;</div>
            <h4>Smart Search</h4>
            <p>Search across thousands of routes with real-time availability</p>
          </div>
          <div className="feature-item">
            <div className="feature-icon">&#128176;</div>
            <h4>Best Prices</h4>
            <p>Get discounts up to 30% on connecting flights</p>
          </div>
          <div className="feature-item">
            <div className="feature-icon">&#128274;</div>
            <h4>Secure Payments</h4>
            <p>Industry-standard payment processing with Stripe</p>
          </div>
          <div className="feature-item">
            <div className="feature-icon">&#9889;</div>
            <h4>Instant Booking</h4>
            <p>Book and get confirmation in seconds</p>
          </div>
        </div>
      </section>
    </div>
  );
}
