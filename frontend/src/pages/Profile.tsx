import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { profileApi, ApiError } from '../api/client';
import { useAuth } from '../context/AuthContext';
import './Profile.css';

export function Profile() {
  const { isLoggedIn, refreshUser } = useAuth();
  const navigate = useNavigate();
  const [fullName, setFullName] = useState('');
  const [phone, setPhone] = useState('');
  const [dob, setDob] = useState('');
  const [gender, setGender] = useState('');
  const [address, setAddress] = useState('');
  const [email, setEmail] = useState('');
  const [saving, setSaving] = useState(false);
  const [saved, setSaved] = useState(false);
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (!isLoggedIn) { navigate('/login'); return; }
    profileApi.get().then(u => {
      setFullName(u.full_name || '');
      setPhone(u.phone || '');
      setDob(u.date_of_birth || '');
      setGender(u.gender || '');
      setAddress(u.address || '');
      setEmail(u.email || '');
    }).catch(() => {}).finally(() => setLoading(false));
  }, [isLoggedIn, navigate]);

  async function handleSave(e: React.FormEvent) {
    e.preventDefault();
    setError('');
    setSaving(true);
    setSaved(false);
    try {
      await profileApi.update({
        full_name: fullName,
        phone,
        date_of_birth: dob || undefined,
        gender: gender || undefined,
        address: address || undefined,
      });
      setSaved(true);
      refreshUser();
      setTimeout(() => setSaved(false), 3000);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Failed to save');
    } finally {
      setSaving(false);
    }
  }

  if (loading) return <div className="flights-loading"><span className="spinner" /></div>;

  return (
    <div className="profile-page">
      <div className="profile-card">
        <h1>Profile</h1>
        <p className="profile-email">{email}</p>

        {error && <div className="auth-error">{error}</div>}
        {saved && <div className="profile-saved">Profile saved successfully!</div>}

        <form onSubmit={handleSave}>
          <div className="form-group">
            <label>Full Name</label>
            <input value={fullName} onChange={e => setFullName(e.target.value)} placeholder="John Doe" />
          </div>
          <div className="form-group">
            <label>Phone</label>
            <input type="tel" value={phone} onChange={e => setPhone(e.target.value)} placeholder="+91 9876543210" />
          </div>
          <div className="form-row-profile">
            <div className="form-group">
              <label>Date of Birth</label>
              <input type="date" value={dob} onChange={e => setDob(e.target.value)} />
            </div>
            <div className="form-group">
              <label>Gender</label>
              <select value={gender} onChange={e => setGender(e.target.value)}>
                <option value="">Select</option>
                <option value="male">Male</option>
                <option value="female">Female</option>
                <option value="other">Other</option>
              </select>
            </div>
          </div>
          <div className="form-group">
            <label>Address</label>
            <textarea value={address} onChange={e => setAddress(e.target.value)} rows={3} placeholder="Your address" />
          </div>
          <button type="submit" className="btn btn-primary auth-submit" disabled={saving}>
            {saving ? <span className="spinner" /> : 'Save Profile'}
          </button>
        </form>
      </div>
    </div>
  );
}
