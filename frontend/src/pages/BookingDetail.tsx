import { useState, useEffect } from 'react';
import { useParams, useSearchParams, Link, useNavigate } from 'react-router-dom';
import { bookingsApi, flightsApi, type Booking, type Flight } from '../api/client';
import { useAuth } from '../context/AuthContext';
import './BookingDetail.css';

export function BookingDetail() {
  const { id } = useParams<{ id: string }>();
  const [searchParams, setSearchParams] = useSearchParams();
  const { isLoggedIn, loading: authLoading } = useAuth();
  const navigate = useNavigate();
  const [booking, setBooking] = useState<Booking | null>(null);
  const [flight, setFlight] = useState<Flight | null>(null);
  const [loading, setLoading] = useState(true);
  const [confirming, setConfirming] = useState(false);

  useEffect(() => {
    // Wait for auth to finish loading before checking login status
    if (authLoading) return;
    if (!isLoggedIn) { navigate('/login'); return; }
    if (!id) return;

    const paymentStatus = searchParams.get('payment');
    const sessionId = searchParams.get('session_id') || '';

    bookingsApi.get(id).then(async (b) => {
      setBooking(b);
      flightsApi.get(b.flight_id).then(setFlight).catch(() => {});

      // Auto-confirm after Stripe Checkout redirect
      if (paymentStatus === 'success' && b.status === 'pending') {
        setConfirming(true);
        try {
          const confirmed = await bookingsApi.confirmByBookingId(id, sessionId);
          setBooking(confirmed);
        } catch {
          // Payment may still be processing — booking will be confirmed by webhook
        } finally {
          setConfirming(false);
          setSearchParams({}, { replace: true });
        }
      }
    }).catch(() => {}).finally(() => setLoading(false));
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [id, isLoggedIn, authLoading, navigate]);

  if (authLoading || loading || confirming) {
    return (
      <div className="flights-loading">
        <span className="spinner" />
        {confirming && <p style={{ marginTop: '1rem', color: 'var(--text-muted)' }}>Confirming your payment...</p>}
      </div>
    );
  }

  if (!booking) return <div className="no-flights">Booking not found.</div>;

  const statusClass = booking.status === 'confirmed' ? 'status-confirmed' : booking.status === 'pending' ? 'status-pending' : 'status-other';

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
            <div className="detail-row"><span>Departure</span><span>{new Date(flight.departure_time).toLocaleString()}</span></div>
            <div className="detail-row"><span>Arrival</span><span>{new Date(flight.arrival_time).toLocaleString()}</span></div>
            <div className="detail-row"><span>Price per Seat</span><span>${(flight.price / 100).toFixed(2)}</span></div>
          </div>
        )}

        <div className="booking-actions">
          <Link to="/flights" className="btn btn-secondary">Search More Flights</Link>
          <Link to="/bookings" className="btn btn-primary">My Bookings</Link>
        </div>
      </div>
    </div>
  );
}
