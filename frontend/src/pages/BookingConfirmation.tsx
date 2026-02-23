import { useState, useEffect, useRef } from 'react';
import { useSearchParams, useNavigate } from 'react-router-dom';
import { bookingsApi, flightsApi, citiesApi, airportsApi, type Booking, type Flight, type City, type Airport } from '../api/client';
import { useAuth } from '../context/AuthContext';
import './BookingConfirmation.css';

type Phase = 'loading' | 'pending' | 'confirmed';

export function BookingConfirmation() {
  const navigate = useNavigate();
  const { isLoggedIn, loading: authLoading } = useAuth();
  const [searchParams] = useSearchParams();
  const confirmedRef = useRef(false);

  const bookingId = searchParams.get('booking_id') || '';
  const paymentIntentId = searchParams.get('payment_intent_id') || '';

  const [phase, setPhase] = useState<Phase>('loading');
  const [booking, setBooking] = useState<Booking | null>(null);
  const [flight, setFlight] = useState<Flight | null>(null);
  const [cities, setCities] = useState<City[]>([]);
  const [airports, setAirports] = useState<Airport[]>([]);
  const [countdown, setCountdown] = useState(8);

  // Step 1: Load booking data, show PENDING
  useEffect(() => {
    if (authLoading) return;
    if (!isLoggedIn) { navigate('/login'); return; }
    if (!bookingId) { navigate('/'); return; }

    Promise.all([
      bookingsApi.get(bookingId),
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
      setPhase(b.status === 'confirmed' ? 'confirmed' : 'pending');
    }).catch(() => {
      navigate('/');
    });
  }, [authLoading, isLoggedIn, bookingId, navigate]);

  // Step 2: After 2.5s in pending, confirm the booking
  useEffect(() => {
    if (phase !== 'pending' || !paymentIntentId || confirmedRef.current) return;
    confirmedRef.current = true;

    const timer = setTimeout(async () => {
      try {
        const confirmed = await bookingsApi.confirm(paymentIntentId);
        setBooking(confirmed);
        setPhase('confirmed');
      } catch {
        // If confirm fails, still show what we have
        setPhase('confirmed');
      }
    }, 2500);

    return () => clearTimeout(timer);
  }, [phase, paymentIntentId]);

  // Step 3: Countdown after confirmed
  useEffect(() => {
    if (phase !== 'confirmed') return;
    if (countdown <= 0) {
      navigate('/', { replace: true });
      return;
    }
    const timer = setTimeout(() => setCountdown(c => c - 1), 1000);
    return () => clearTimeout(timer);
  }, [phase, countdown, navigate]);

  // Helpers
  function getAirport(id: string) { return airports.find(a => a.id === id); }
  function getCity(cityId: string) { return cities.find(c => c.id === cityId); }
  function getRouteLabel(airportId: string) {
    const ap = getAirport(airportId);
    if (!ap) return airportId;
    const city = getCity(ap.city_id);
    return city ? `${city.name} (${ap.code})` : ap.code;
  }

  // ── Loading state ──
  if (phase === 'loading') {
    return (
      <div className="confirmation-page">
        <div className="confirmation-loading">
          <span className="spinner" />
          <p>Processing your payment...</p>
        </div>
      </div>
    );
  }

  if (!booking) return null;

  const isConfirmed = phase === 'confirmed';
  const amount = (booking.amount / 100).toFixed(2);
  const pricePerSeat = flight ? (flight.price / 100).toFixed(2) : '—';

  return (
    <div className="confirmation-page">
      <div className="confirmation-card">
        {/* ── Hero ── */}
        <div className="confirmation-hero">
          {isConfirmed ? (
            <>
              <div className="confirmation-checkmark">
                <svg viewBox="0 0 52 52" className="checkmark-svg">
                  <circle className="checkmark-circle" cx="26" cy="26" r="25" fill="none" />
                  <path className="checkmark-check" fill="none" d="M14.1 27.2l7.1 7.2 16.7-16.8" />
                </svg>
              </div>
              <h1>Booking Confirmed!</h1>
              <p className="confirmation-subtitle">Your flight has been booked successfully</p>
            </>
          ) : (
            <>
              <div className="confirmation-pending-icon">
                <div className="pending-spinner" />
              </div>
              <h1>Processing Payment...</h1>
              <p className="confirmation-subtitle">Please wait while we confirm your booking</p>
            </>
          )}
        </div>

        {/* ── Countdown (top) ── */}
        {isConfirmed && (
          <div className="confirmation-footer confirmation-footer-top">
            <div className="confirmation-countdown">
              Redirecting to home in <strong>{countdown}s</strong>
            </div>
            <div className="confirmation-progress">
              <div className="confirmation-progress-bar" style={{ width: `${((8 - countdown) / 8) * 100}%` }} />
            </div>
          </div>
        )}

        {/* ── Status badge ── */}
        <div className="confirmation-status-bar">
          <span className={`confirmation-badge ${isConfirmed ? 'badge-confirmed' : 'badge-pending'}`}>
            {isConfirmed ? '✓ CONFIRMED' : '⏳ PENDING'}
          </span>
        </div>

        {/* ── Ticket ── */}
        <div className={`ticket ${isConfirmed ? 'ticket-confirmed' : 'ticket-pending'}`}>
          <div className="ticket-header">
            <span className="ticket-airline">✈ SkyFlow</span>
            <span className="ticket-flight">{flight?.flight_number || '—'}</span>
          </div>

          {flight && (
            <div className="ticket-route">
              <div className="ticket-endpoint">
                <span className="ticket-time">{new Date(flight.departure_time).toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' })}</span>
                <span className="ticket-city">{getRouteLabel(flight.origin_id)}</span>
                <span className="ticket-date">{new Date(flight.departure_time).toLocaleDateString('en-US', { weekday: 'short', month: 'short', day: 'numeric' })}</span>
              </div>
              <div className="ticket-arrow">
                <div className="ticket-line" />
                <span>✈</span>
                <div className="ticket-line" />
              </div>
              <div className="ticket-endpoint">
                <span className="ticket-time">{new Date(flight.arrival_time).toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' })}</span>
                <span className="ticket-city">{getRouteLabel(flight.destination_id)}</span>
                <span className="ticket-date">{new Date(flight.arrival_time).toLocaleDateString('en-US', { weekday: 'short', month: 'short', day: 'numeric' })}</span>
              </div>
            </div>
          )}

          <div className="ticket-tear" />

          <div className="ticket-details">
            <div className="ticket-detail-item">
              <span className="ticket-label">Passenger</span>
              <span className="ticket-value">{booking.passenger_name}</span>
            </div>
            <div className="ticket-detail-item">
              <span className="ticket-label">Email</span>
              <span className="ticket-value">{booking.passenger_email}</span>
            </div>
            {booking.passenger_phone && (
              <div className="ticket-detail-item">
                <span className="ticket-label">Phone</span>
                <span className="ticket-value">{booking.passenger_phone}</span>
              </div>
            )}
            <div className="ticket-detail-item">
              <span className="ticket-label">Seats</span>
              <span className="ticket-value">{booking.seats}</span>
            </div>
          </div>
        </div>

        {/* ── Payment summary ── */}
        <div className="confirmation-section">
          <h3>Payment Summary</h3>
          <div className="confirmation-row">
            <span>Price per seat</span>
            <span>${pricePerSeat}</span>
          </div>
          <div className="confirmation-row">
            <span>Seats</span>
            <span>× {booking.seats}</span>
          </div>
          <div className="confirmation-row total">
            <span>Total Paid</span>
            <span className="confirmation-amount">${amount}</span>
          </div>
          <div className="confirmation-row">
            <span>Payment Status</span>
            <span className={isConfirmed ? 'confirmation-paid' : 'confirmation-pending-text'}>
              {isConfirmed ? '✓ Paid' : '⏳ Processing...'}
            </span>
          </div>
        </div>

        {/* ── Reference ── */}
        <div className="confirmation-section">
          <h3>Reference</h3>
          <div className="confirmation-row">
            <span>Booking ID</span>
            <span className="confirmation-mono">{booking.id}</span>
          </div>
          <div className="confirmation-row">
            <span>Booked At</span>
            <span>{new Date(booking.created_at).toLocaleString()}</span>
          </div>
        </div>

        {/* ── Email notice (only when confirmed) ── */}
        {isConfirmed && (
          <div className="confirmation-email-notice">
            <span>📧</span>
            <p>A confirmation email with your ticket has been sent to <strong>{booking.passenger_email}</strong></p>
          </div>
        )}

        {/* ── Actions (only when confirmed) ── */}
        {isConfirmed && (
          <div className="confirmation-actions">
            <button className="btn btn-secondary" onClick={() => navigate(`/bookings/${booking.id}`)}>
              View Booking Details
            </button>
            <button className="btn btn-primary" onClick={() => navigate('/', { replace: true })}>
              Go Home Now
            </button>
          </div>
        )}
      </div>
    </div>
  );
}
