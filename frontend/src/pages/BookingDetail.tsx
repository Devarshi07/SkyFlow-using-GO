import { useState, useEffect, useRef } from 'react';
import { useParams, useSearchParams, Link, useNavigate } from 'react-router-dom';
import { bookingsApi, flightsApi, citiesApi, airportsApi, type Booking, type Flight, type City, type Airport, ApiError } from '../api/client';
import { gqlApi } from '../api/graphql';
import { useAuth } from '../context/AuthContext';
import './BookingDetail.css';

export function BookingDetail() {
  const { id } = useParams<{ id: string }>();
  const [searchParams, setSearchParams] = useSearchParams();
  const { isLoggedIn, loading: authLoading } = useAuth();
  const navigate = useNavigate();
  const confirmedRef = useRef(false);

  const justPaid = searchParams.get('just_paid') === 'true';
  const paymentIntentId = searchParams.get('pi') || '';
  const paymentStatus = searchParams.get('payment');
  const sessionId = searchParams.get('session_id') || '';

  // Edit-payment flow (Stripe redirect back)
  const editPaid = searchParams.get('edit_paid') === 'true';
  const editSessionId = searchParams.get('session_id') || '';
  const editNewFlightId = searchParams.get('new_flight_id') || '';
  const editNewSeats = parseInt(searchParams.get('new_seats') || '0', 10);

  const [booking, setBooking] = useState<Booking | null>(null);
  const [flight, setFlight] = useState<Flight | null>(null);
  const [cities, setCities] = useState<City[]>([]);
  const [airports, setAirports] = useState<Airport[]>([]);
  const [loading, setLoading] = useState(true);
  const [editOpen, setEditOpen] = useState(false);

  // Post-payment flow state
  const [postPaymentPhase, setPostPaymentPhase] = useState<'none' | 'pending' | 'confirmed'>('none');
  const [countdown, setCountdown] = useState(8);

  // Fetch data
  useEffect(() => {
    if (authLoading) return;
    if (!isLoggedIn) { navigate('/login'); return; }
    if (!id) return;

    Promise.all([
      bookingsApi.get(id),
      citiesApi.list(),
      airportsApi.list(),
    ]).then(async ([b, c, a]) => {
      setBooking(b);
      setCities(c);
      setAirports(a);
      try {
        const f = await flightsApi.get(b.flight_id);
        setFlight(f);
      } catch { /* ignore */ }

      // Determine if this is a post-payment flow
      if (editPaid) {
        // Edit payment completed via Stripe — confirm the edit
        setPostPaymentPhase('pending');
      } else if (justPaid && b.status === 'pending') {
        setPostPaymentPhase('pending');
      } else if (paymentStatus === 'success' && b.status === 'pending') {
        setPostPaymentPhase('pending');
      } else if (b.status === 'confirmed' && (justPaid || paymentStatus === 'success')) {
        // Already confirmed (e.g. webhook beat us)
        setPostPaymentPhase('confirmed');
      }
    }).catch(() => {}).finally(() => setLoading(false));
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [id, isLoggedIn, authLoading, navigate]);

  // After 2.5s in pending, confirm the booking
  useEffect(() => {
    if (postPaymentPhase !== 'pending' || confirmedRef.current) return;
    confirmedRef.current = true;

    const timer = setTimeout(async () => {
      try {
        let confirmed: Booking;
        if (editPaid && id) {
          // Confirm the edit with Stripe session
          confirmed = await bookingsApi.confirmEdit(id, {
            payment_intent_id: '',
            session_id: editSessionId,
            new_flight_id: editNewFlightId,
            new_seats: editNewSeats,
          });
          // Refresh flight data since it may have changed
          try {
            const f = await flightsApi.get(confirmed.flight_id);
            setFlight(f);
          } catch { /* ignore */ }
        } else if (paymentIntentId) {
          confirmed = await bookingsApi.confirm(paymentIntentId);
        } else if (sessionId && id) {
          confirmed = await bookingsApi.confirmByBookingId(id, sessionId);
        } else if (id) {
          confirmed = await bookingsApi.confirmByBookingId(id, '');
        } else {
          setPostPaymentPhase('confirmed');
          return;
        }
        setBooking(confirmed);
      } catch {
        // Still transition visually
      }
      setPostPaymentPhase('confirmed');
      // Clean URL params
      setSearchParams({}, { replace: true });
    }, 2500);

    return () => clearTimeout(timer);
  }, [postPaymentPhase, paymentIntentId, sessionId, id, setSearchParams]);

  // Countdown after confirmed
  useEffect(() => {
    if (postPaymentPhase !== 'confirmed') return;
    if (countdown <= 0) {
      navigate('/', { replace: true });
      return;
    }
    const timer = setTimeout(() => setCountdown(c => c - 1), 1000);
    return () => clearTimeout(timer);
  }, [postPaymentPhase, countdown, navigate]);

  // Helpers
  function getAirport(aid: string) { return airports.find(a => a.id === aid); }
  function getCity(cid: string) { return cities.find(c => c.id === cid); }
  function getRouteLabel(airportId: string) {
    const ap = getAirport(airportId);
    if (!ap) return '';
    const city = getCity(ap.city_id);
    return city ? `${city.name} (${ap.code})` : ap.code;
  }

  if (authLoading || loading) {
    return (
      <div className="flights-loading">
        <span className="spinner" />
      </div>
    );
  }

  if (!booking) return <div className="no-flights">Booking not found.</div>;

  const isPostPayment = postPaymentPhase !== 'none';
  const isConfirmed = postPaymentPhase === 'confirmed' || booking.status === 'confirmed';
  const isPending = postPaymentPhase === 'pending';
  const amount = (booking.amount / 100).toFixed(2);
  const pricePerSeat = flight ? (flight.price / 100).toFixed(2) : '—';

  // ═══════════════════════════════════════════
  // POST-PAYMENT EXPERIENCE
  // ═══════════════════════════════════════════
  if (isPostPayment) {
    return (
      <div className="booking-detail-page">
        <div className="booking-detail-card post-payment-card">
          {/* Hero */}
          <div className="pp-hero">
            {isConfirmed ? (
              <>
                <div className="pp-checkmark">
                  <svg viewBox="0 0 52 52" className="pp-checkmark-svg">
                    <circle className="pp-circle" cx="26" cy="26" r="25" fill="none" />
                    <path className="pp-check" fill="none" d="M14.1 27.2l7.1 7.2 16.7-16.8" />
                  </svg>
                </div>
                <h1>Booking Confirmed!</h1>
                <p className="pp-sub">Your flight has been booked successfully</p>
              </>
            ) : (
              <>
                <div className="pp-pending-spinner-wrap">
                  <div className="pp-pending-spinner" />
                </div>
                <h1>Processing Payment...</h1>
                <p className="pp-sub">Please wait while we confirm your booking</p>
              </>
            )}
          </div>

          {/* Status badge */}
          <div className="pp-badge-bar">
            <span className={`pp-badge ${isConfirmed ? 'pp-badge-confirmed' : 'pp-badge-pending'}`}>
              {isConfirmed ? '✓ CONFIRMED' : '⏳ PENDING'}
            </span>
          </div>

          {/* Countdown (right below badge) */}
          {isConfirmed && (
            <div className="pp-countdown-top">
              <div className="pp-countdown">
                Redirecting to home in <strong>{countdown}s</strong>
              </div>
              <div className="pp-progress">
                <div className="pp-progress-bar" style={{ width: `${((8 - countdown) / 8) * 100}%` }} />
              </div>
            </div>
          )}

          {/* Ticket */}
          <div className={`pp-ticket ${isConfirmed ? 'pp-ticket-confirmed' : 'pp-ticket-pending'}`}>
            <div className="pp-ticket-header">
              <span className="pp-ticket-airline">✈ SkyFlow</span>
              <span className="pp-ticket-flight">{flight?.flight_number || '—'}</span>
            </div>

            {flight && (
              <div className="pp-ticket-route">
                <div className="pp-endpoint">
                  <span className="pp-time">{new Date(flight.departure_time).toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' })}</span>
                  <span className="pp-city">{getRouteLabel(flight.origin_id)}</span>
                  <span className="pp-date">{new Date(flight.departure_time).toLocaleDateString('en-US', { weekday: 'short', month: 'short', day: 'numeric' })}</span>
                </div>
                <div className="pp-arrow">
                  <div className="pp-line" />
                  <span>✈</span>
                  <div className="pp-line" />
                </div>
                <div className="pp-endpoint">
                  <span className="pp-time">{new Date(flight.arrival_time).toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' })}</span>
                  <span className="pp-city">{getRouteLabel(flight.destination_id)}</span>
                  <span className="pp-date">{new Date(flight.arrival_time).toLocaleDateString('en-US', { weekday: 'short', month: 'short', day: 'numeric' })}</span>
                </div>
              </div>
            )}

            <div className="pp-tear" />

            <div className="pp-ticket-details">
              <div className="pp-detail"><span className="pp-label">Passenger</span><span className="pp-val">{booking.passenger_name}</span></div>
              <div className="pp-detail"><span className="pp-label">Email</span><span className="pp-val">{booking.passenger_email}</span></div>
              {booking.passenger_phone && <div className="pp-detail"><span className="pp-label">Phone</span><span className="pp-val">{booking.passenger_phone}</span></div>}
              <div className="pp-detail"><span className="pp-label">Seats</span><span className="pp-val">{booking.seats}</span></div>
            </div>
          </div>

          {/* Payment summary */}
          <div className="pp-section">
            <h3>Payment Summary</h3>
            <div className="pp-row"><span>Price per seat</span><span>${pricePerSeat}</span></div>
            <div className="pp-row"><span>Seats</span><span>× {booking.seats}</span></div>
            <div className="pp-row pp-row-total"><span>Total Paid</span><span className="pp-amount">${amount}</span></div>
            <div className="pp-row">
              <span>Payment Status</span>
              <span className={isConfirmed ? 'pp-paid' : 'pp-processing'}>{isConfirmed ? '✓ Paid' : '⏳ Processing...'}</span>
            </div>
          </div>

          {/* Reference */}
          <div className="pp-section">
            <h3>Reference</h3>
            <div className="pp-row"><span>Booking ID</span><span className="pp-mono">{booking.id}</span></div>
            <div className="pp-row"><span>Booked At</span><span>{new Date(booking.created_at).toLocaleString()}</span></div>
          </div>

          {/* Email notice + actions (only when confirmed) */}
          {isConfirmed && (
            <>
              <div className="pp-email-notice">
                <span>📧</span>
                <p>A confirmation email has been sent to <strong>{booking.passenger_email}</strong></p>
              </div>

              <div className="pp-actions">
                <button className="btn btn-secondary" onClick={() => { setPostPaymentPhase('none'); setSearchParams({}, { replace: true }); }}>
                  View Booking Details
                </button>
                <button className="btn btn-primary" onClick={() => navigate('/', { replace: true })}>
                  Go Home Now
                </button>
              </div>
            </>
          )}
        </div>
      </div>
    );
  }

  // ═══════════════════════════════════════════
  // NORMAL BOOKING DETAIL VIEW
  // ═══════════════════════════════════════════
  const statusClass = booking.status === 'confirmed' ? 'status-confirmed' : booking.status === 'pending' ? 'status-pending' : 'status-other';
  const canEdit = booking.status !== 'cancelled';

  return (
    <div className="booking-detail-page">
      <div className="booking-detail-card">
        {booking.status === 'confirmed' && (
          <div className="payment-success-banner">
            <span className="success-icon">✓</span>
            <div>
              <strong>Payment Confirmed</strong>
              <p>Your booking has been confirmed successfully.</p>
            </div>
          </div>
        )}

        <div className="booking-header">
          <h1>Booking Confirmation</h1>
          <span className={`status-badge ${statusClass}`}>{booking.status.toUpperCase()}</span>
        </div>

        <div className="booking-section">
          <h3>Booking Details</h3>
          <div className="detail-row"><span>Booking ID</span><span className="mono">{booking.id}</span></div>
          <div className="detail-row"><span>Passenger</span><span>{booking.passenger_name}</span></div>
          <div className="detail-row"><span>Email</span><span>{booking.passenger_email}</span></div>
          {booking.passenger_phone && <div className="detail-row"><span>Phone</span><span>{booking.passenger_phone}</span></div>}
          <div className="detail-row"><span>Seats</span><span>{booking.seats}</span></div>
          <div className="detail-row"><span>Booked At</span><span>{new Date(booking.created_at).toLocaleString()}</span></div>
        </div>

        {flight && (
          <div className="booking-section">
            <h3>Flight Details</h3>
            <div className="detail-row"><span>Flight</span><span>{flight.flight_number}</span></div>
            {getRouteLabel(flight.origin_id) && (
              <div className="detail-row"><span>From</span><span>{getRouteLabel(flight.origin_id)}</span></div>
            )}
            {getRouteLabel(flight.destination_id) && (
              <div className="detail-row"><span>To</span><span>{getRouteLabel(flight.destination_id)}</span></div>
            )}
            <div className="detail-row"><span>Departure</span><span>{new Date(flight.departure_time).toLocaleString()}</span></div>
            <div className="detail-row"><span>Arrival</span><span>{new Date(flight.arrival_time).toLocaleString()}</span></div>
            <div className="detail-row"><span>Price per Seat</span><span>${(flight.price / 100).toFixed(2)}</span></div>
          </div>
        )}

        {booking.amount > 0 && (
          <div className="booking-section">
            <h3>Payment</h3>
            <div className="detail-row"><span>Total</span><span>${amount}</span></div>
            <div className="detail-row"><span>Status</span><span className={booking.status === 'confirmed' ? 'status-text-confirmed' : ''}>{booking.status === 'confirmed' ? '✓ Paid' : booking.status}</span></div>
          </div>
        )}

        <div className="booking-actions">
          {canEdit && (
            <button className="btn btn-secondary" onClick={() => setEditOpen(true)}>Edit Booking</button>
          )}
          <Link to="/flights" className="btn btn-secondary">Search More Flights</Link>
          <Link to="/bookings" className="btn btn-primary">My Bookings</Link>
        </div>
      </div>

      {editOpen && booking && (
        <EditBookingModal
          booking={booking}
          flight={flight}
          onClose={() => setEditOpen(false)}
          onSaved={(updated) => {
            setBooking(updated);
            setEditOpen(false);
            if (updated.flight_id !== booking.flight_id) {
              flightsApi.get(updated.flight_id).then(setFlight).catch(() => {});
            }
          }}
        />
      )}
    </div>
  );
}

type EditTab = 'passenger' | 'flight';

function EditBookingModal({ booking, flight: currentFlight, onClose, onSaved }: {
  booking: Booking;
  flight: Flight | null;
  onClose: () => void;
  onSaved: (b: Booking) => void;
}) {
  const navigate = useNavigate();
  const [tab, setTab] = useState<EditTab>('passenger');
  const [error, setError] = useState('');
  const [saving, setSaving] = useState(false);

  // Passenger fields
  const [name, setName] = useState(booking.passenger_name);
  const [email, setEmail] = useState(booking.passenger_email);
  const [phone, setPhone] = useState(booking.passenger_phone || '');
  const [seats, setSeats] = useState(booking.seats);

  // Flight change fields
  const [allAirports, setAllAirports] = useState<Airport[]>([]);
  const [allCities, setAllCities] = useState<City[]>([]);
  const [origin, setOrigin] = useState('');
  const [dest, setDest] = useState('');
  const [date, setDate] = useState(new Date().toISOString().split('T')[0]);
  const [searchResults, setSearchResults] = useState<Flight[]>([]);
  const [searching, setSearching] = useState(false);
  const [searched, setSearched] = useState(false);
  const [selectedFlightId, setSelectedFlightId] = useState<string | null>(null);

  // Load airports/cities for dropdowns
  useEffect(() => {
    gqlApi.airportsAndCities().then(d => {
      setAllAirports(d.airports || []);
      setAllCities(d.cities || []);
    }).catch(() => {});
  }, []);

  function cityName(cityId: string) {
    return allCities.find(c => c.id === cityId)?.name || '';
  }

  async function handleFlightSearch() {
    if (!origin || !dest || !date) return;
    setSearching(true);
    setSearched(true);
    setSelectedFlightId(null);
    try {
      const res = await gqlApi.searchFlights(origin, dest, date);
      setSearchResults(res.searchFlights.flights || []);
    } catch {
      setSearchResults([]);
    } finally {
      setSearching(false);
    }
  }

  // Price diff calculation
  const selectedFlight = searchResults.find(f => f.id === selectedFlightId);
  const currentAmount = booking.amount;
  const newAmount = selectedFlight
    ? selectedFlight.price * (seats > 0 ? seats : booking.seats)
    : currentFlight
      ? currentFlight.price * seats
      : currentAmount;
  const priceDiff = selectedFlightId && selectedFlightId !== booking.flight_id
    ? newAmount - currentAmount
    : (seats !== booking.seats && currentFlight)
      ? (currentFlight.price * seats) - currentAmount
      : 0;

  async function handleSave() {
    setError('');
    setSaving(true);
    try {
      const payload: Record<string, unknown> = {
        passenger_name: name,
        passenger_email: email,
        passenger_phone: phone,
        seats,
      };
      if (selectedFlightId && selectedFlightId !== booking.flight_id) {
        payload.flight_id = selectedFlightId;
      }
      const res = await bookingsApi.edit(booking.id, payload as any);

      if (res.needs_payment) {
        if (res.checkout_url) {
          // Stripe Checkout — redirect to Stripe's hosted page
          window.location.href = res.checkout_url;
          return;
        }
        // Demo mode — use built-in checkout page
        const params = new URLSearchParams();
        params.set('booking_id', booking.id);
        params.set('payment_intent_id', res.payment_intent_id || '');
        params.set('amount', (res.amount_due || 0).toString());
        params.set('flight_id', selectedFlightId || booking.flight_id);
        params.set('name', name);
        params.set('email', email);
        params.set('edit_confirm', 'true');
        params.set('new_flight_id', selectedFlightId || '');
        params.set('new_seats', seats.toString());
        navigate(`/checkout?${params.toString()}`);
        return;
      }

      onSaved(res.booking);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Failed to update booking');
    } finally {
      setSaving(false);
    }
  }

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal-content edit-modal-wide" onClick={e => e.stopPropagation()}>
        <h2>Edit Booking</h2>
        <p style={{ color: 'var(--text-muted)', marginBottom: '1rem', fontSize: '0.85rem' }}>
          Booking ID: {booking.id.slice(0, 8)}...
        </p>

        {/* Tabs */}
        <div className="edit-tabs">
          <button className={`edit-tab ${tab === 'passenger' ? 'active' : ''}`} onClick={() => setTab('passenger')}>
            Passenger Details
          </button>
          <button className={`edit-tab ${tab === 'flight' ? 'active' : ''}`} onClick={() => setTab('flight')}>
            Change Flight
          </button>
        </div>

        {error && <div className="auth-error">{error}</div>}

        {/* Passenger tab */}
        {tab === 'passenger' && (
          <div className="edit-tab-content">
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
              <input type="number" min={1} max={10} value={seats} onChange={e => setSeats(Number(e.target.value))} />
            </div>
          </div>
        )}

        {/* Flight tab */}
        {tab === 'flight' && (
          <div className="edit-tab-content">
            <div className="edit-flight-search">
              <div className="form-group">
                <label>From</label>
                <select value={origin} onChange={e => setOrigin(e.target.value)}>
                  <option value="">Select origin</option>
                  {allAirports.map(a => (
                    <option key={a.id} value={a.id}>{a.code} — {cityName(a.city_id)}</option>
                  ))}
                </select>
              </div>
              <div className="form-group">
                <label>To</label>
                <select value={dest} onChange={e => setDest(e.target.value)}>
                  <option value="">Select destination</option>
                  {allAirports.filter(a => a.id !== origin).map(a => (
                    <option key={a.id} value={a.id}>{a.code} — {cityName(a.city_id)}</option>
                  ))}
                </select>
              </div>
              <div className="form-group">
                <label>Date</label>
                <input type="date" value={date} onChange={e => setDate(e.target.value)} />
              </div>
              <button
                type="button"
                className="btn btn-primary"
                style={{ width: '100%', marginTop: '0.25rem' }}
                onClick={handleFlightSearch}
                disabled={!origin || !dest || !date || searching}
              >
                {searching ? <span className="spinner" /> : 'Search Flights'}
              </button>
            </div>

            {/* Search results */}
            {searched && !searching && searchResults.length === 0 && (
              <p style={{ color: 'var(--text-muted)', textAlign: 'center', margin: '1rem 0', fontSize: '0.85rem' }}>
                No flights found for this route and date.
              </p>
            )}

            {searchResults.length > 0 && (
              <div className="edit-flight-results">
                {searchResults.map(f => {
                  const isCurrent = f.id === booking.flight_id;
                  const isSelected = f.id === selectedFlightId;
                  return (
                    <div
                      key={f.id}
                      className={`edit-flight-option ${isSelected ? 'selected' : ''} ${isCurrent ? 'current' : ''}`}
                      onClick={() => !isCurrent && setSelectedFlightId(f.id)}
                    >
                      <div className="edit-flight-option-top">
                        <span className="edit-flight-num">{f.flight_number}</span>
                        {isCurrent && <span className="edit-flight-current-badge">Current</span>}
                        {isSelected && <span className="edit-flight-selected-badge">Selected</span>}
                        <span className="edit-flight-price">${(f.price / 100).toFixed(2)}</span>
                      </div>
                      <div className="edit-flight-option-bottom">
                        <span>{new Date(f.departure_time).toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' })}</span>
                        <span style={{ color: 'var(--text-muted)' }}>→</span>
                        <span>{new Date(f.arrival_time).toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' })}</span>
                        <span style={{ color: 'var(--text-muted)', marginLeft: 'auto', fontSize: '0.8rem' }}>{f.seats_available} seats</span>
                      </div>
                    </div>
                  );
                })}
              </div>
            )}
          </div>
        )}

        {/* Price difference notice */}
        {priceDiff !== 0 && (
          <div className={`edit-price-diff ${priceDiff > 0 ? 'diff-increase' : 'diff-decrease'}`}>
            {priceDiff > 0
              ? `Additional payment of ${(priceDiff / 100).toFixed(2)} required`
              : `You save ${(Math.abs(priceDiff) / 100).toFixed(2)} with this change`
            }
          </div>
        )}

        {/* Save / Cancel */}
        <div style={{ display: 'flex', gap: '0.75rem', marginTop: '1.25rem' }}>
          <button type="button" className="btn btn-secondary" onClick={onClose} style={{ flex: 1 }}>Cancel</button>
          <button type="button" className="btn btn-primary" disabled={saving} onClick={handleSave} style={{ flex: 1 }}>
            {saving
              ? <span className="spinner" />
              : priceDiff > 0
                ? `Pay ${(priceDiff / 100).toFixed(2)} & Save`
                : 'Save Changes'
            }
          </button>
        </div>
      </div>
    </div>
  );
}
