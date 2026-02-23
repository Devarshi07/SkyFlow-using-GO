import { useState, useEffect } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { bookingsApi, type Booking } from '../api/client';
import { useAuth } from '../context/AuthContext';
import './MyBookings.css';

export function MyBookings() {
  const { isLoggedIn, loading: authLoading } = useAuth();
  const navigate = useNavigate();
  const [bookings, setBookings] = useState<Booking[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (authLoading) return;
    if (!isLoggedIn) { navigate('/login'); return; }
    bookingsApi.my().then(setBookings).catch(() => {}).finally(() => setLoading(false));
  }, [isLoggedIn, authLoading, navigate]);

  if (authLoading || loading) return <div className="flights-loading"><span className="spinner" /></div>;

  return (
    <div className="my-bookings-page">
      <h1>My Bookings</h1>
      {bookings.length === 0 ? (
        <div className="no-bookings">
          <p>You don't have any bookings yet.</p>
          <Link to="/flights" className="btn btn-primary">Search Flights</Link>
        </div>
      ) : (
        <div className="bookings-list">
          {bookings.map(b => {
            const statusClass = b.status === 'confirmed' ? 'status-confirmed' : b.status === 'pending' ? 'status-pending' : 'status-other';
            return (
              <Link key={b.id} to={`/bookings/${b.id}`} className="booking-item">
                <div className="booking-item-left">
                  <span className="booking-item-name">{b.passenger_name}</span>
                  <span className="booking-item-date">{new Date(b.created_at).toLocaleDateString()}</span>
                </div>
                <div className="booking-item-right">
                  <span className={`status-badge ${statusClass}`}>{b.status}</span>
                  <span className="booking-item-seats">{b.seats} seat{b.seats > 1 ? 's' : ''}</span>
                </div>
              </Link>
            );
          })}
        </div>
      )}
    </div>
  );
}
