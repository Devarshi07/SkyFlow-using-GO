import { useState, useEffect } from 'react';
import { Link, useNavigate, useSearchParams } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import { ApiError } from '../api/client';
import './Auth.css';

// Module-level guard survives StrictMode remounts AND re-renders
let oauthInFlight = false;

export function Login() {
  const { login, googleLogin } = useAuth();
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();

  // Where to go after login
  const redirectTo = searchParams.get('redirect') || localStorage.getItem('login_redirect') || '/';

  // Persist redirect so Register page can use it too
  useEffect(() => {
    const r = searchParams.get('redirect');
    if (r) localStorage.setItem('login_redirect', r);
  }, [searchParams]);
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const [googleLoading, setGoogleLoading] = useState(false);

  // Handle OAuth callback: exchange code for tokens when redirected back from Google
  useEffect(() => {
    const code = searchParams.get('code');
    if (!code || oauthInFlight) return;

    // Set module-level guard immediately
    oauthInFlight = true;

    // Strip the code from the URL right away so it can't be reused
    setSearchParams({}, { replace: true });

    const redirectUri = window.location.origin + '/login';
    setGoogleLoading(true);
    setError('');

    googleLogin(code, redirectUri)
      .then(() => {
        const dest = localStorage.getItem('login_redirect') || '/';
        localStorage.removeItem('login_redirect');
        navigate(dest, { replace: true });
      })
      .catch((err) => {
        setError(err instanceof ApiError ? err.message : 'Google sign-in failed');
      })
      .finally(() => {
        oauthInFlight = false;
        setGoogleLoading(false);
      });
    // Only depend on searchParams — googleLogin is stable via useCallback now
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [searchParams]);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError('');
    setLoading(true);
    try {
      await login(email, password);
      localStorage.removeItem('login_redirect');
      navigate(redirectTo, { replace: true });
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Login failed');
    } finally {
      setLoading(false);
    }
  }

  function handleGoogleLogin() {
    // Save redirect URL before leaving for Google OAuth
    if (redirectTo && redirectTo !== '/') {
      localStorage.setItem('login_redirect', redirectTo);
    }
    const clientId = import.meta.env.VITE_GOOGLE_CLIENT_ID || '1013267431983-nb49iiaav62n32jnb7186u0k5m5p1e5p.apps.googleusercontent.com';
    const redirectUri = window.location.origin + '/login';
    const scope = 'openid email profile';
    const url = `https://accounts.google.com/o/oauth2/v2/auth?client_id=${clientId}&redirect_uri=${encodeURIComponent(redirectUri)}&response_type=code&scope=${encodeURIComponent(scope)}&access_type=offline&prompt=consent`;
    window.location.href = url;
  }

  if (googleLoading) {
    return (
      <div className="auth-page">
        <div className="auth-card">
          <h1>Sign In</h1>
          <p className="auth-subtitle">Completing sign in with Google…</p>
          <div className="auth-loading"><span className="spinner" /></div>
        </div>
      </div>
    );
  }

  return (
    <div className="auth-page">
      <div className="auth-card">
        <h1>Sign In</h1>
        <p className="auth-subtitle">Welcome back to SkyFlow</p>

        {error && <div className="auth-error">{error}</div>}

        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <label htmlFor="email">Email</label>
            <input id="email" type="email" value={email} onChange={e => setEmail(e.target.value)} placeholder="you@example.com" required />
          </div>
          <div className="form-group">
            <label htmlFor="password">Password</label>
            <input id="password" type="password" value={password} onChange={e => setPassword(e.target.value)} placeholder="••••••" required />
          </div>
          <button type="submit" className="btn btn-primary auth-submit" disabled={loading}>
            {loading ? <span className="spinner" /> : 'Sign In'}
          </button>
        </form>

        <div className="auth-divider"><span>or</span></div>

        <button className="btn btn-secondary google-btn" onClick={handleGoogleLogin}>
          <svg width="18" height="18" viewBox="0 0 48 48"><path fill="#4285F4" d="M24 9.5c3.5 0 6.6 1.2 9.1 3.6l6.8-6.8C35.8 2.5 30.3 0 24 0 14.6 0 6.7 5.6 2.7 13.6l7.9 6.1C12.5 13.4 17.8 9.5 24 9.5z"/><path fill="#34A853" d="M46.1 24.5c0-1.7-.1-3.3-.4-4.9H24v9.3h12.4c-.5 2.8-2.1 5.2-4.5 6.8l7 5.4c4.1-3.8 6.4-9.3 6.4-16.6z"/><path fill="#FBBC05" d="M10.6 28.3a14.5 14.5 0 010-8.6l-7.9-6.1a24 24 0 000 20.8l7.9-6.1z"/><path fill="#EA4335" d="M24 48c6.3 0 11.6-2.1 15.5-5.7l-7-5.4c-2.2 1.5-5.1 2.4-8.5 2.4-6.2 0-11.5-3.9-13.4-9.5l-7.9 6.1C6.7 42.4 14.6 48 24 48z"/></svg>
          Sign in with Google
        </button>

        <p className="auth-footer">
          <Link to="/forgot-password">Forgot your password?</Link>
        </p>
        <p className="auth-footer">
          Don't have an account? <Link to="/register">Create one</Link>
        </p>
      </div>
    </div>
  );
}
