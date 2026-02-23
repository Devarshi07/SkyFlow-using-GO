import { useState, useEffect } from 'react';
import { useParams, useSearchParams, Link, useNavigate } from 'react-router-dom';
import { bookingsApi, flightsApi, type Booking, type Flight, ApiError } from '../api/client';
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
  const [editOpen, setEditOpen] = useState(false);

  useEffect(() => {
    if (authLoading) return;
    if (!isLoggedIn) { navigate('/login'); return; }
    if (!id) return;

    const paymentStatus = searchParams.get('payment');
    const sessionId = searchParams.get('session_id') || '';

    bookingsApi.get(id).then(async (b) => {
      setBooking(b);
      flightsApi.get(b.flight_id).then(setFlight).catch(() => {});

      if (paymentStatus === 'success' && b.status === 'pending') {
        setConfirming(true);
        try {
          await bookingsApi.confirmByBookingId(id, sessionId);
          navigate(`/booking-confirmed?booking_id=${id}`, { replace: true });
          return;
        } catch {
          // Payment may still be processing
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
            <div className="detail-row"><span>Departure</span><span>{new Date(flight.departure_time).toLocaleString()}</span></div>
            <div className="detail-row"><span>Arrival</span><span>{new Date(flight.arrival_time).toLocaleString()}</span></div>
            <div className="detail-row"><span>Price per Seat</span><span>${(flight.price / 100).toFixed(2)}</span></div>
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

function EditBookingModal({ booking, onClose, onSaved }: {
  booking: Booking;
  onClose: () => void;
  onSaved: (b: Booking) => void;
}) {
  const [name, setName] = useState(booking.passenger_name);
  const [email, setEmail] = useState(booking.passenger_email);
  const [phone, setPhone] = useState(booking.passenger_phone || '');
  const [seats, setSeats] = useState(booking.seats);
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError('');
    setLoading(true);
    try {
      const updated = await bookingsApi.edit(booking.id, {
        passenger_name: name,
        passenger_email: email,
        passenger_phone: phone,
        seats,
      });
      onSaved(updated);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Failed to update booking');
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal-content" onClick={e => e.stopPropagation()}>
        <h2>Edit Booking</h2>
        <p style={{ color: 'var(--text-muted)', marginBottom: '1rem', fontSize: '0.85rem' }}>
          Booking ID: {booking.id.slice(0, 8)}...
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
            <input type="number" min={1} max={10} value={seats} onChange={e => setSeats(Number(e.target.value))} />
          </div>
          <div style={{ display: 'flex', gap: '0.75rem', marginTop: '1rem' }}>
            <button type="button" className="btn btn-secondary" onClick={onClose} style={{ flex: 1 }}>Cancel</button>
            <button type="submit" className="btn btn-primary" disabled={loading} style={{ flex: 1 }}>
              {loading ? <span className="spinner" /> : 'Save Changes'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
