import { useState } from 'react';
import { Link, useSearchParams, useNavigate } from 'react-router-dom';
import { authApi, ApiError } from '../api/client';
import './Auth.css';

export function ResetPassword() {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const token = searchParams.get('token') || '';
  const [password, setPassword] = useState('');
  const [confirm, setConfirm] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const [success, setSuccess] = useState(false);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (password !== confirm) {
      setError('Passwords do not match');
      return;
    }
    if (!token) {
      setError('Invalid reset link');
      return;
    }
    setError('');
    setLoading(true);
    try {
      await authApi.resetPassword(token, password);
      setSuccess(true);
      setTimeout(() => navigate('/login'), 3000);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Failed to reset password');
    } finally {
      setLoading(false);
    }
  }

  if (success) {
    return (
      <div className="auth-page">
        <div className="auth-card">
          <h1>Password Reset!</h1>
          <p className="auth-subtitle">
            Your password has been updated successfully. Redirecting to sign in...
          </p>
          <Link to="/login" className="btn btn-primary auth-submit">Go to Sign In</Link>
        </div>
      </div>
    );
  }

  if (!token) {
    return (
      <div className="auth-page">
        <div className="auth-card">
          <h1>Invalid Reset Link</h1>
          <p className="auth-subtitle">This password reset link is invalid or has expired.</p>
          <Link to="/forgot-password" className="btn btn-primary auth-submit">Request New Link</Link>
        </div>
      </div>
    );
  }

  return (
    <div className="auth-page">
      <div className="auth-card">
        <h1>Set New Password</h1>
        <p className="auth-subtitle">Enter your new password below.</p>

        {error && <div className="auth-error">{error}</div>}

        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <label htmlFor="password">New Password</label>
            <input id="password" type="password" value={password} onChange={e => setPassword(e.target.value)} placeholder="Min 6 characters" required minLength={6} />
          </div>
          <div className="form-group">
            <label htmlFor="confirm">Confirm Password</label>
            <input id="confirm" type="password" value={confirm} onChange={e => setConfirm(e.target.value)} placeholder="••••••" required />
          </div>
          <button type="submit" className="btn btn-primary auth-submit" disabled={loading}>
            {loading ? <span className="spinner" /> : 'Reset Password'}
          </button>
        </form>

        <p className="auth-footer">
          <Link to="/login">Back to Sign In</Link>
        </p>
      </div>
    </div>
  );
}
