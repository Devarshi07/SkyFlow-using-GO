import { useState, useEffect } from 'react';
import { useSearchParams, useNavigate } from 'react-router-dom';
import { bookingsApi, flightsApi, citiesApi, airportsApi, type Flight, type Airport, type City, ApiError } from '../api/client';
import { SkyFlowLogo } from '../components/SkyFlowLogo';
import { useAuth } from '../context/AuthContext';
import './ReviewBooking.css';

function formatTime(iso: string) {
  return new Date(iso).toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit', hour12: true });
}
function formatDate(iso: string) {
  return new Date(iso).toLocaleDateString('en-US', { weekday: 'long', month: 'short', day: 'numeric', year: 'numeric' });
}
function formatDuration(dep: string, arr: string) {
  const min = (new Date(arr).getTime() - new Date(dep).getTime()) / 60000;
  const h = Math.floor(min / 60);
  const m = Math.round(min % 60);
  return `${h}h ${m}m`;
}

export function ReviewBooking() {
  const navigate = useNavigate();
  const { isLoggedIn, loading: authLoading } = useAuth();
  const [searchParams] = useSearchParams();

  const flightId = searchParams.get('flight_id') || '';
  const seats = parseInt(searchParams.get('seats') || '1', 10);
  const name = searchParams.get('name') || '';
  const email = searchParams.get('email') || '';
  const phone = searchParams.get('phone') || '';

  const [flight, setFlight] = useState<Flight | null>(null);
  const [airports, setAirports] = useState<Airport[]>([]);
  const [cities, setCities] = useState<City[]>([]);
  const [loading, setLoading] = useState(true);
  const [booking, setBooking] = useState(false);
  const [error, setError] = useState('');

  useEffect(() => {
    if (authLoading) return;
    if (!isLoggedIn) { navigate('/login'); return; }
    if (!flightId) { navigate('/flights'); return; }

    Promise.all([
      flightsApi.get(flightId),
      airportsApi.list(),
      citiesApi.list(),
    ]).then(([f, a, c]) => {
      setFlight(f);
      setAirports(a);
      setCities(c);
    }).catch(() => navigate('/flights'))
      .finally(() => setLoading(false));
  }, [authLoading, isLoggedIn, flightId, navigate]);

  function getAirport(id: string) { return airports.find(a => a.id === id); }
  function getCity(cityId: string) { return cities.find(c => c.id === cityId); }
  function airportLabel(id: string) {
    const ap = getAirport(id);
    if (!ap) return '';
    const city = getCity(ap.city_id);
    return city ? `${city.name} (${ap.code})` : ap.code;
  }
  function airportName(id: string) {
    const ap = getAirport(id);
    return ap?.name || '';
  }
  function cityName(id: string) {
    const ap = getAirport(id);
    if (!ap) return '';
    return getCity(ap.city_id)?.name || '';
  }

  async function handleProceed() {
    if (!flight) return;
    setError('');
    setBooking(true);
    try {
      const res = await bookingsApi.create({
        flight_id: flight.id,
        seats,
        passenger_name: name,
        passenger_email: email,
        passenger_phone: phone,
      });
      if (res.checkout_url) {
        window.location.href = res.checkout_url;
        return;
      }
      if (res.payment_intent_id && res.payment_intent_id.startsWith('pi_mock')) {
        const params = new URLSearchParams();
        params.set('booking_id', res.booking_id);
        params.set('payment_intent_id', res.payment_intent_id);
        params.set('amount', res.amount.toString());
        params.set('flight_id', flight.id);
        params.set('name', name);
        params.set('email', email);
        navigate(`/checkout?${params.toString()}`);
        return;
      }
      if (res.payment_intent_id) {
        setError('Payment setup failed — please try again');
        setBooking(false);
        return;
      }
      navigate(`/bookings/${res.booking_id}`);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Booking failed');
      setBooking(false);
    }
  }

  if (authLoading || loading) {
    return <div className="review-loading"><span className="spinner" /> Loading flight details...</div>;
  }
  if (!flight) return null;

  const baseFare = flight.price * seats;
  const taxes = Math.round(baseFare * 0.12);
  const total = baseFare + taxes;

  return (
    <div className="review-page">
      {/* Header */}
      <div className="review-header">
        <h1>Review your booking</h1>
        <p className="review-sub">{cityName(flight.origin_id)} → {cityName(flight.destination_id)}</p>
      </div>

      <div className="review-layout">
        {/* Left: Itinerary */}
        <div className="review-main">
          {/* Route summary */}
          <div className="review-card">
            <div className="review-route-header">
              <div>
                <h2>{cityName(flight.origin_id)} → {cityName(flight.destination_id)}</h2>
                <span className="review-route-date">{formatDate(flight.departure_time)}</span>
                <span className="review-route-meta"> · Non Stop · {formatDuration(flight.departure_time, flight.arrival_time)}</span>
              </div>
            </div>

            {/* Airline */}
            <div className="review-airline-row">
              <SkyFlowLogo size="sm" flightNumbers={flight.flight_number} />
              <div className="review-class-badge">Economy</div>
            </div>

            {/* Timeline */}
            <div className="review-timeline">
              <div className="timeline-stop">
                <div className="timeline-time">{formatTime(flight.departure_time)}</div>
                <div className="timeline-dot" />
                <div className="timeline-info">
                  <div className="timeline-city">{cityName(flight.origin_id)}</div>
                  <div className="timeline-airport">{airportName(flight.origin_id)}</div>
                </div>
              </div>

              <div className="timeline-duration">
                <div className="timeline-line" />
                <span className="timeline-duration-text">{formatDuration(flight.departure_time, flight.arrival_time)}</span>
              </div>

              <div className="timeline-stop">
                <div className="timeline-time">{formatTime(flight.arrival_time)}</div>
                <div className="timeline-dot" />
                <div className="timeline-info">
                  <div className="timeline-city">{cityName(flight.destination_id)}</div>
                  <div className="timeline-airport">{airportName(flight.destination_id)}</div>
                </div>
              </div>
            </div>
          </div>

          {/* Traveller Details */}
          <div className="review-card">
            <h3 className="review-card-title">Traveller Details</h3>
            <div className="review-traveller">
              <div className="traveller-row">
                <span className="traveller-label">Passenger</span>
                <span className="traveller-value">{name}</span>
              </div>
              <div className="traveller-row">
                <span className="traveller-label">Email</span>
                <span className="traveller-value">{email}</span>
              </div>
              <div className="traveller-row">
                <span className="traveller-label">Phone</span>
                <span className="traveller-value">{phone}</span>
              </div>
              <div className="traveller-row">
                <span className="traveller-label">Seats</span>
                <span className="traveller-value">{seats} {seats === 1 ? 'Adult' : 'Adults'}</span>
              </div>
            </div>
          </div>
        </div>

        {/* Right: Fare Summary */}
        <div className="review-sidebar">
          <div className="review-card fare-card">
            <h3 className="review-card-title">Fare Summary</h3>

            <div className="fare-row">
              <span>Base Fare ({seats} {seats === 1 ? 'traveller' : 'travellers'})</span>
              <span>${(baseFare / 100).toFixed(2)}</span>
            </div>
            <div className="fare-row">
              <span>Taxes & Surcharges</span>
              <span>${(taxes / 100).toFixed(2)}</span>
            </div>

            <div className="fare-divider" />

            <div className="fare-row fare-total">
              <span>Total Amount</span>
              <span>${(total / 100).toFixed(2)}</span>
            </div>

            {error && <div className="auth-error" style={{ marginTop: '1rem' }}>{error}</div>}

            <button
              className="btn btn-primary review-pay-btn"
              onClick={handleProceed}
              disabled={booking}
            >
              {booking ? <span className="spinner" /> : `Proceed to Payment — $${(total / 100).toFixed(2)}`}
            </button>

            <p className="review-secure">🔒 Secured by SkyFlow Payments</p>
          </div>

          <button className="btn btn-secondary review-back-btn" onClick={() => navigate(-1)}>
            ← Back to Flights
          </button>
        </div>
      </div>
    </div>
  );
}
