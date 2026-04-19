import { useState, useRef, useEffect } from 'react';
import { Link, Outlet, useNavigate, useLocation } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import { ChatWidget } from './ChatWidget';
import './Layout.css';

export function Layout() {
  const { isLoggedIn, user, logout } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();
  const [dropdownOpen, setDropdownOpen] = useState(false);
  const [navHidden, setNavHidden] = useState(false);
  const dropdownRef = useRef<HTMLDivElement>(null);
  const lastScrollY = useRef(0);
  const ticking = useRef(false);

  useEffect(() => {
    function handleClick(e: MouseEvent) {
      if (dropdownRef.current && !dropdownRef.current.contains(e.target as Node)) {
        setDropdownOpen(false);
      }
    }
    document.addEventListener('mousedown', handleClick);
    return () => document.removeEventListener('mousedown', handleClick);
  }, []);

  useEffect(() => {
    function onScroll() {
      const y = window.scrollY;
      if (y < 10) {
        setNavHidden(false);
      } else if (y > lastScrollY.current + 60) {
        setNavHidden(true);
      } else if (y < lastScrollY.current - 30) {
        setNavHidden(false);
      }
      lastScrollY.current = y;
      ticking.current = false;
    }
    function requestTick() {
      if (!ticking.current) {
        requestAnimationFrame(() => {
          onScroll();
        });
        ticking.current = true;
      }
    }
    window.addEventListener('scroll', requestTick, { passive: true });
    return () => window.removeEventListener('scroll', requestTick);
  }, []);

  const displayName = user?.full_name || user?.email?.split('@')[0] || '';
  const initial = (displayName[0] || '?').toUpperCase();
  const isHome = location.pathname === '/';

  function handleLogout() {
    logout();
    setDropdownOpen(false);
    navigate('/');
  }

  return (
    <>
      <header className={`navbar ${navHidden ? 'navbar-hidden' : ''}`}>
        <div className="navbar-inner container">
          <Link to="/" className="navbar-brand">
            <span className="brand-icon">&#9992;</span>
            <span className="brand-text">SkyFlow</span>
          </Link>

          <nav className="navbar-links">
            <Link to="/" className={`nav-link ${location.pathname === '/' ? 'active' : ''}`}>
              <span className="nav-icon">&#9992;</span>
              Flights
            </Link>
            {isLoggedIn && (
              <Link to="/bookings" className={`nav-link ${location.pathname === '/bookings' ? 'active' : ''}`}>
                <span className="nav-icon">&#128203;</span>
                My Trips
              </Link>
            )}
          </nav>

          <div className="navbar-right">
            <span className="navbar-currency">USD</span>
            {isLoggedIn ? (
              <div className="avatar-wrap" ref={dropdownRef}>
                <button className="avatar-btn" onClick={() => setDropdownOpen(!dropdownOpen)}>
                  <span className="avatar-circle">{initial}</span>
                  <span className="avatar-name">{displayName}</span>
                  <svg className="avatar-chevron" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5"><path d="M6 9l6 6 6-6"/></svg>
                </button>
                {dropdownOpen && (
                  <div className="avatar-dropdown">
                    <div className="dropdown-header">
                      <span className="dropdown-avatar">{initial}</span>
                      <div>
                        <div className="dropdown-name">{displayName}</div>
                        <div className="dropdown-email">{user?.email}</div>
                      </div>
                    </div>
                    <div className="dropdown-divider" />
                    <Link to="/profile" className="dropdown-item" onClick={() => setDropdownOpen(false)}>
                      <span className="dropdown-item-icon">&#128100;</span>
                      Profile
                    </Link>
                    <Link to="/bookings" className="dropdown-item" onClick={() => setDropdownOpen(false)}>
                      <span className="dropdown-item-icon">&#128203;</span>
                      My Bookings
                    </Link>
                    <div className="dropdown-divider" />
                    <button className="dropdown-item dropdown-logout" onClick={handleLogout}>
                      <span className="dropdown-item-icon">&#128682;</span>
                      Logout
                    </button>
                  </div>
                )}
              </div>
            ) : (
              <div className="auth-buttons">
                <Link to="/login" className="login-link">Login</Link>
                <Link to="/register" className="register-btn">Create Account</Link>
              </div>
            )}
          </div>
        </div>
      </header>
      <main className={isHome ? 'page-flush' : 'page'}>
        <div className={isHome ? '' : 'container'}>
          <Outlet />
        </div>
      </main>

      <ChatWidget />
    </>
  );
}
