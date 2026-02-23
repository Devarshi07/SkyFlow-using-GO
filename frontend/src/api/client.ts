const API = '/api/v1';

function getToken(): string | null {
  return localStorage.getItem('access_token');
}

function setTokens(access: string, refresh: string) {
  localStorage.setItem('access_token', access);
  localStorage.setItem('refresh_token', refresh);
}

function clearTokens() {
  localStorage.removeItem('access_token');
  localStorage.removeItem('refresh_token');
}

// erasableSyntaxOnly: true forbids parameter properties (public x),
// so declare fields explicitly
export class ApiError extends Error {
  status: number;
  code?: string;

  constructor(status: number, message: string, code?: string) {
    super(message);
    this.status = status;
    this.code = code;
  }
}

async function request<T>(path: string, opts: RequestInit = {}): Promise<T> {
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...(opts.headers as Record<string, string> || {}),
  };
  const token = getToken();
  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }

  let res = await fetch(`${API}${path}`, { ...opts, headers });

  if (res.status === 401 && token) {
    const refreshed = await tryRefresh();
    if (refreshed) {
      headers['Authorization'] = `Bearer ${getToken()}`;
      res = await fetch(`${API}${path}`, { ...opts, headers });
    }
  }

  if (res.status === 204) return undefined as T;

  const data = await res.json();
  if (!res.ok) {
    // Surface detailed error when backend returns details (DEBUG=1)
    let msg = data.message || 'Request failed';
    if (data.details?.cause) {
      msg = `${msg}: ${data.details.cause}`;
    }
    throw new ApiError(res.status, msg, data.code);
  }
  return data as T;
}

async function tryRefresh(): Promise<boolean> {
  const rt = localStorage.getItem('refresh_token');
  if (!rt) return false;
  try {
    const res = await fetch(`${API}/auth/refresh`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ refresh_token: rt }),
    });
    if (!res.ok) {
      clearTokens();
      return false;
    }
    const data = await res.json();
    setTokens(data.access_token, data.refresh_token);
    return true;
  } catch (_err: unknown) {
    clearTokens();
    return false;
  }
}

// --- Types ---

export interface TokenResponse {
  access_token: string;
  refresh_token: string;
  expires_in: number;
  token_type: string;
}

export interface User {
  id: string;
  email: string;
  full_name?: string;
  phone?: string;
  date_of_birth?: string;
  gender?: string;
  address?: string;
  created_at: string;
}

export interface Flight {
  id: string;
  flight_number: string;
  origin_id: string;
  destination_id: string;
  departure_time: string;
  arrival_time: string;
  price: number;
  seats_total: number;
  seats_available: number;
}

export interface ConnectingFlight {
  leg1: Flight;
  leg2: Flight;
  total_price: number;
  discount: number;
  total_duration_hours: number;
  layover_minutes: number;
  layover_airport_id: string;
}

export interface SearchResult {
  flights: Flight[];
  connecting: ConnectingFlight[];
}

export interface Airport {
  id: string;
  name: string;
  city_id: string;
  code: string;
}

export interface City {
  id: string;
  name: string;
  country: string;
  code: string;
}

export interface Booking {
  id: string;
  user_id: string;
  flight_id: string;
  seats: number;
  passenger_name: string;
  passenger_email: string;
  passenger_phone?: string;
  payment_intent_id?: string;
  status: string;
  created_at: string;
}

export interface EditBookingResponse {
  booking: Booking;
  needs_payment: boolean;
  payment_intent_id?: string;
  amount_due?: number;
  old_amount?: number;
  new_amount?: number;
}

export interface CreateBookingResponse {
  booking_id: string;
  payment_intent_id?: string;
  checkout_url?: string;
  amount: number;
  status: string;
}

// --- API namespaces ---

export const authApi = {
  login: (email: string, password: string) =>
    request<TokenResponse>('/auth/login', {
      method: 'POST',
      body: JSON.stringify({ email, password }),
    }).then(r => { setTokens(r.access_token, r.refresh_token); return r; }),

  register: (email: string, password: string) =>
    request<TokenResponse>('/auth/register', {
      method: 'POST',
      body: JSON.stringify({ email, password }),
    }).then(r => { setTokens(r.access_token, r.refresh_token); return r; }),

  googleAuth: (code: string, redirectUri: string) =>
    request<TokenResponse>('/auth/google', {
      method: 'POST',
      body: JSON.stringify({ code, redirect_uri: redirectUri }),
    }).then(r => { setTokens(r.access_token, r.refresh_token); return r; }),

  me: () => request<User>('/auth/me'),

  forgotPassword: (email: string) =>
    request<{ message: string }>('/auth/forgot-password', {
      method: 'POST',
      body: JSON.stringify({ email }),
    }),

  resetPassword: (token: string, newPassword: string) =>
    request<{ message: string }>('/auth/reset-password', {
      method: 'POST',
      body: JSON.stringify({ token, new_password: newPassword }),
    }),

  logout: () => {
    const rt = localStorage.getItem('refresh_token');
    clearTokens();
    return request<void>('/auth/logout', {
      method: 'POST',
      body: JSON.stringify({ refresh_token: rt }),
    }).catch(() => { /* ignore logout errors */ });
  },
};

export const profileApi = {
  get: () => request<User>('/auth/me'),
  update: (data: Partial<Pick<User, 'full_name' | 'phone' | 'date_of_birth' | 'gender' | 'address'>>) =>
    request<User>('/auth/profile', {
      method: 'PUT',
      body: JSON.stringify(data),
    }),
};

export const flightsApi = {
  search: (originId: string, destId: string, date: string) =>
    request<SearchResult>(`/flights/search?origin_id=${originId}&destination_id=${destId}&date=${date}`),
  get: (id: string) => request<Flight>(`/flights/${id}`),
};

export const airportsApi = {
  list: () => request<{ airports: Airport[] }>('/airports').then(r => r.airports),
};

export const citiesApi = {
  list: () => request<{ cities: City[] }>('/cities').then(r => r.cities),
};

export const bookingsApi = {
  create: (data: { flight_id: string; seats: number; passenger_name: string; passenger_email: string; passenger_phone?: string }) =>
    request<CreateBookingResponse>('/bookings', {
      method: 'POST',
      body: JSON.stringify(data),
    }),
  confirm: (paymentIntentId: string) =>
    request<Booking>('/bookings/confirm', {
      method: 'POST',
      body: JSON.stringify({ payment_intent_id: paymentIntentId }),
    }),
  confirmByBookingId: (bookingId: string, sessionId?: string) =>
    request<Booking>(`/bookings/${bookingId}/confirm-payment`, {
      method: 'POST',
      body: JSON.stringify({ session_id: sessionId || '' }),
    }),
  cancel: (bookingId: string) =>
    request<Booking>(`/bookings/${bookingId}/cancel`, {
      method: 'POST',
    }),
  edit: (bookingId: string, data: { flight_id?: string; seats?: number; passenger_name?: string; passenger_email?: string; passenger_phone?: string }) =>
    request<EditBookingResponse>(`/bookings/${bookingId}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    }),
  confirmEdit: (bookingId: string, data: { payment_intent_id: string; new_flight_id: string; new_seats: number }) =>
    request<Booking>(`/bookings/${bookingId}/confirm-edit`, {
      method: 'POST',
      body: JSON.stringify(data),
    }),
  get: (id: string) => request<Booking>(`/bookings/${id}`),
  my: () => request<{ bookings: Booking[] }>('/bookings/my').then(r => r.bookings),
};

export { getToken, clearTokens, setTokens };
