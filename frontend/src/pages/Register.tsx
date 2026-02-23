import { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import { ApiError } from '../api/client';
import './Auth.css';

function validatePassword(pw: string) {
  return {
    minLength: pw.length >= 8,
    hasUpper: /[A-Z]/.test(pw),
    hasLower: /[a-z]/.test(pw),
    hasNumber: /\d/.test(pw),
  };
}

export function Register() {
  const { register } = useAuth();
  const navigate = useNavigate();
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [confirm, setConfirm] = useState('');
  const [phone, setPhone] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  const pwReqs = validatePassword(password);
  const pwValid = pwReqs.minLength && pwReqs.hasUpper && pwReqs.hasLower && pwReqs.hasNumber;

  function formatPhone(v: string) {
    return v.replace(/[^\d+\-() ]/g, '').slice(0, 20);
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!pwValid) {
      setError('Password does not meet requirements');
      return;
    }
    if (password !== confirm) {
      setError('Passwords do not match');
      return;
    }
    setError('');
    setLoading(true);
    try {
      await register(email, password);
      const dest = localStorage.getItem('login_redirect') || '/';
      localStorage.removeItem('login_redirect');
      navigate(dest, { replace: true });
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Registration failed');
    } finally {
      setLoading(false);
    }
  }

  function handleGoogleSignup() {
    const clientId = import.meta.env.VITE_GOOGLE_CLIENT_ID || '1013267431983-nb49iiaav62n32jnb7186u0k5m5p1e5p.apps.googleusercontent.com';
    const redirectUri = window.location.origin + '/login';
    const scope = 'openid email profile';
    const url = `https://accounts.google.com/o/oauth2/v2/auth?client_id=${clientId}&redirect_uri=${encodeURIComponent(redirectUri)}&response_type=code&scope=${encodeURIComponent(scope)}&access_type=offline&prompt=consent`;
    window.location.href = url;
  }

  return (
    <div className="auth-page">
      <div className="auth-card">
        <h1>Create Account</h1>
        <p className="auth-subtitle">Join SkyFlow and start booking flights</p>

        {error && <div className="auth-error">{error}</div>}

        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <label htmlFor="email">Email</label>
            <input id="email" type="email" value={email} onChange={e => setEmail(e.target.value)} placeholder="you@example.com" required />
          </div>
          <div className="form-group">
            <label htmlFor="phone">Phone Number</label>
            <input id="phone" type="tel" value={phone} onChange={e => setPhone(formatPhone(e.target.value))} placeholder="+1 (555) 000-0000" />
          </div>
          <div className="form-group">
            <label htmlFor="password">Password</label>
            <input id="password" type="password" value={password} onChange={e => setPassword(e.target.value)} placeholder="Create a strong password" required />
            {password.length > 0 && (
              <div className="password-requirements">
                <div className={pwReqs.minLength ? 'req-met' : 'req-unmet'}>{pwReqs.minLength ? '✓' : '○'} At least 8 characters</div>
                <div className={pwReqs.hasUpper ? 'req-met' : 'req-unmet'}>{pwReqs.hasUpper ? '✓' : '○'} One uppercase letter</div>
                <div className={pwReqs.hasLower ? 'req-met' : 'req-unmet'}>{pwReqs.hasLower ? '✓' : '○'} One lowercase letter</div>
                <div className={pwReqs.hasNumber ? 'req-met' : 'req-unmet'}>{pwReqs.hasNumber ? '✓' : '○'} One number</div>
              </div>
            )}
          </div>
          <div className="form-group">
            <label htmlFor="confirm">Confirm Password</label>
            <input id="confirm" type="password" value={confirm} onChange={e => setConfirm(e.target.value)} placeholder="••••••••" required />
            {confirm.length > 0 && confirm !== password && (
              <div className="field-error">Passwords do not match</div>
            )}
          </div>
          <button type="submit" className="btn btn-primary auth-submit" disabled={loading || !pwValid}>
            {loading ? <span className="spinner" /> : 'Create Account'}
          </button>
        </form>

        <div className="auth-divider"><span>or</span></div>

        <button className="btn btn-secondary google-btn" onClick={handleGoogleSignup}>
          <svg width="18" height="18" viewBox="0 0 48 48"><path fill="#4285F4" d="M24 9.5c3.5 0 6.6 1.2 9.1 3.6l6.8-6.8C35.8 2.5 30.3 0 24 0 14.6 0 6.7 5.6 2.7 13.6l7.9 6.1C12.5 13.4 17.8 9.5 24 9.5z"/><path fill="#34A853" d="M46.1 24.5c0-1.7-.1-3.3-.4-4.9H24v9.3h12.4c-.5 2.8-2.1 5.2-4.5 6.8l7 5.4c4.1-3.8 6.4-9.3 6.4-16.6z"/><path fill="#FBBC05" d="M10.6 28.3a14.5 14.5 0 010-8.6l-7.9-6.1a24 24 0 000 20.8l7.9-6.1z"/><path fill="#EA4335" d="M24 48c6.3 0 11.6-2.1 15.5-5.7l-7-5.4c-2.2 1.5-5.1 2.4-8.5 2.4-6.2 0-11.5-3.9-13.4-9.5l-7.9 6.1C6.7 42.4 14.6 48 24 48z"/></svg>
          Sign up with Google
        </button>

        <p className="auth-footer">
          Already have an account? <Link to="/login">Sign in</Link>
        </p>
      </div>
    </div>
  );
}
