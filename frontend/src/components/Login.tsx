// @ts-nocheck
// Login page for device code (YubiKey) authentication
import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';

const Login: React.FC = () => {
  const [deviceCode, setDeviceCode] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const navigate = useNavigate();

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError('');
    try {
      const res = await fetch('/api/v1/auth/session', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          device_type: 'yubikey',
          auth_code: deviceCode,
          permission: 'yubiapp:read',
        }),
      });
      const data = await res.json();
      if (!res.ok) {
        setError(data.error || 'Login failed.');
        setLoading(false);
        return;
      }
      // Store tokens in localStorage (or context, as needed)
      localStorage.setItem('access_token', data.access_token);
      localStorage.setItem('refresh_token', data.refresh_token);
      localStorage.setItem('session_id', data.session_id);
      navigate('/dashboard');
    } catch (err) {
      setError('Network error.');
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50">
      <div className="w-full max-w-md p-8 bg-white rounded-lg shadow-lg">
        <div className="flex flex-col items-center mb-6">
          <span className="text-2xl font-bold text-blue-700 mb-2">YubiApp Login</span>
          <span className="text-sm text-gray-500">Sign in with your device code</span>
        </div>
        <form onSubmit={handleLogin} className="space-y-4">
          <div>
            <label htmlFor="deviceCode" className="block text-sm font-medium text-gray-700">
              Device Code (YubiKey or other)
            </label>
            <input
              id="deviceCode"
              type="text"
              autoFocus
              required
              value={deviceCode}
              onChange={e => setDeviceCode(e.target.value)}
              className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500"
              placeholder="Touch your YubiKey or enter code"
              disabled={loading}
            />
          </div>
          {error && <div className="text-red-600 text-sm">{error}</div>}
          <button
            type="submit"
            className="w-full py-2 px-4 bg-blue-700 text-white font-semibold rounded-md hover:bg-blue-800 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 disabled:opacity-50"
            disabled={loading}
          >
            {loading ? 'Logging in...' : 'Login'}
          </button>
        </form>
        <div className="mt-6 text-center text-xs text-gray-400">
          &copy; {new Date().getFullYear()} Fanjango Limited
        </div>
      </div>
    </div>
  );
};

export default Login; 