import { useState, useEffect, useMemo } from 'react';
import { useSearchParams, useNavigate } from 'react-router-dom';
import { bookingsApi, flightsApi, type Flight, ApiError } from '../api/client';
import { useAuth } from '../context/AuthContext';
import './Checkout.css';

function formatCard(v: string) {
  const digits = v.replace(/\D/g, '').slice(0, 16);
  return digits.replace(/(.{4})/g, '$1 ').trim();
}

function formatExpiry(v: string) {
  const digits = v.replace(/\D/g, '').slice(0, 4);
  if (digits.length >= 3) return digits.slice(0, 2) + '/' + digits.slice(2);
  return digits;
}

export function Checkout() {
  const { isLoggedIn, loading: authLoading } = useAuth();
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();

  const bookingId = searchParams.get('booking_id') || '';
  const paymentIntentId = searchParams.get('payment_intent_id') || '';
  const amountParam = parseInt(searchParams.get('amount') || '0', 10);
  const flightId = searchParams.get('flight_id') || '';
  const passengerName = searchParams.get('name') || '';
  const passengerEmail = searchParams.get('email') || '';

  // Edit-confirm flow
  const isEditConfirm = searchParams.get('edit_confirm') === 'true';
  const newFlightId = searchParams.get('new_flight_id') || '';
  const newSeats = parseInt(searchParams.get('new_seats') || '0', 10);

  const [flight, setFlight] = useState<Flight | null>(null);
  const [cardNumber, setCardNumber] = useState('');
  const [expiry, setExpiry] = useState('');
  const [cvv, setCvv] = useState('');
  const [cardName, setCardName] = useState(passengerName);
  const [processing, setProcessing] = useState(false);
  const [error, setError] = useState('');

  useEffect(() => {
    if (authLoading) return;
    if (!isLoggedIn) { navigate('/login'); return; }
    if (!bookingId || !paymentIntentId) { navigate('/flights'); return; }
    if (flightId) {
      flightsApi.get(flightId).then(setFlight).catch(() => {});
    }
  }, [authLoading, isLoggedIn, bookingId, paymentIntentId, flightId, navigate]);

  const amountDollars = useMemo(() => (amountParam / 100).toFixed(2), [amountParam]);

  const cardDigits = cardNumber.replace(/\s/g, '');
  const isFormValid = cardDigits.length >= 15 && expiry.length >= 4 && cvv.length >= 3 && cardName.length > 0;

  async function handlePay(e: React.FormEvent) {
    e.preventDefault();
    if (!isFormValid) return;
    setError('');
    setProcessing(true);
    try {
      if (isEditConfirm) {
        // Confirm the edit — pays the diff and applies the flight/seat change
        await bookingsApi.confirmEdit(bookingId, {
          payment_intent_id: paymentIntentId,
          new_flight_id: newFlightId,
          new_seats: newSeats,
        });
        navigate(`/bookings/${bookingId}?just_paid=true&pi=${paymentIntentId}`, { replace: true });
      } else {
        navigate(`/bookings/${bookingId}?just_paid=true&pi=${paymentIntentId}`, { replace: true });
      }
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Payment failed. Please try again.');
      setProcessing(false);
    }
  }

  const cardType = cardDigits.startsWith('4') ? 'visa' :
    cardDigits.startsWith('5') ? 'mastercard' :
    cardDigits.startsWith('3') ? 'amex' : '';

  return (
    <div className="checkout-page">
      <div className="checkout-container">
        <div className="checkout-summary">
          <div className="summary-header">
            <span className="summary-lock">&#128274;</span>
            <h2>{isEditConfirm ? 'Additional Payment' : 'Order Summary'}</h2>
          </div>

          {flight && (
            <div className="summary-flight">
              <div className="summary-flight-number">{flight.flight_number}</div>
              <div className="summary-route">
                <span>{new Date(flight.departure_time).toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' })}</span>
                <span className="summary-arrow">&rarr;</span>
                <span>{new Date(flight.arrival_time).toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' })}</span>
              </div>
              <div className="summary-date">
                {new Date(flight.departure_time).toLocaleDateString('en-US', { weekday: 'long', month: 'long', day: 'numeric', year: 'numeric' })}
              </div>
            </div>
          )}

          <div className="summary-details">
            {passengerName && <div className="summary-row"><span>Passenger</span><span>{passengerName}</span></div>}
            {passengerEmail && <div className="summary-row"><span>Email</span><span>{passengerEmail}</span></div>}
            <div className="summary-row"><span>Booking ID</span><span className="summary-mono">{bookingId.slice(0, 8)}...</span></div>
          </div>

          <div className="summary-total">
            <span>Total</span>
            <span className="summary-amount">${amountDollars}</span>
          </div>

          <div className="summary-secure"><span>&#128274; Secured by SkyFlow Payments</span></div>
        </div>

        <div className="checkout-form-wrap">
          <div className="checkout-form-header">
            <h2>Payment Details</h2>
            <div className="card-brands">
              <span className={`card-brand ${cardType === 'visa' ? 'active' : ''}`}>VISA</span>
              <span className={`card-brand ${cardType === 'mastercard' ? 'active' : ''}`}>MC</span>
              <span className={`card-brand ${cardType === 'amex' ? 'active' : ''}`}>AMEX</span>
            </div>
          </div>

          {error && <div className="checkout-error">{error}</div>}

          <form onSubmit={handlePay} className="checkout-form">
            <div className="form-group">
              <label htmlFor="card-name">Name on Card</label>
              <input id="card-name" type="text" value={cardName} onChange={e => setCardName(e.target.value)} placeholder="John Doe" required autoComplete="cc-name" />
            </div>

            <div className="form-group">
              <label htmlFor="card-number">Card Number</label>
              <input id="card-number" type="text" value={cardNumber} onChange={e => setCardNumber(formatCard(e.target.value))} placeholder="4242 4242 4242 4242" maxLength={19} required autoComplete="cc-number" inputMode="numeric" />
            </div>

            <div className="form-row-checkout">
              <div className="form-group">
                <label htmlFor="expiry">Expiry Date</label>
                <input id="expiry" type="text" value={expiry} onChange={e => setExpiry(formatExpiry(e.target.value))} placeholder="MM/YY" maxLength={5} required autoComplete="cc-exp" inputMode="numeric" />
              </div>
              <div className="form-group">
                <label htmlFor="cvv">CVV</label>
                <input id="cvv" type="password" value={cvv} onChange={e => setCvv(e.target.value.replace(/\D/g, '').slice(0, 4))} placeholder="&bull;&bull;&bull;" maxLength={4} required autoComplete="cc-csc" inputMode="numeric" />
              </div>
            </div>

            <button type="submit" className="btn btn-primary pay-btn" disabled={!isFormValid || processing}>
              {processing ? (
                <span className="pay-processing"><span className="spinner" /> Processing Payment...</span>
              ) : (
                <span>Pay ${amountDollars}</span>
              )}
            </button>

            <p className="checkout-note">
              Demo mode — any card details will be accepted.<br />No real charges will be made.
            </p>
          </form>
        </div>
      </div>
    </div>
  );
}
