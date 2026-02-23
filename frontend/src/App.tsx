import { BrowserRouter, Routes, Route } from 'react-router-dom';
import { AuthProvider } from './context/AuthContext';
import { Layout } from './components/Layout';
import { Home } from './pages/Home';
import { Login } from './pages/Login';
import { Register } from './pages/Register';
import { Flights } from './pages/Flights';
import { Profile } from './pages/Profile';
import { Checkout } from './pages/Checkout';
import { BookingDetail } from './pages/BookingDetail';
import { MyBookings } from './pages/MyBookings';
import './App.css';

export default function App() {
  return (
    <BrowserRouter>
      <AuthProvider>
        <Routes>
          <Route element={<Layout />}>
            <Route path="/" element={<Home />} />
            <Route path="/login" element={<Login />} />
            <Route path="/register" element={<Register />} />
            <Route path="/flights" element={<Flights />} />
            <Route path="/profile" element={<Profile />} />
            <Route path="/checkout" element={<Checkout />} />
            <Route path="/bookings/:id" element={<BookingDetail />} />
            <Route path="/bookings" element={<MyBookings />} />
          </Route>
        </Routes>
      </AuthProvider>
    </BrowserRouter>
  );
}
