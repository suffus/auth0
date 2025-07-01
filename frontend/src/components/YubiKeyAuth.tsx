import React, { useState, useRef, useEffect } from 'react';
import { useAuthenticateDevice, useLogout } from '../hooks/useAuth';
import TopNavBar from './TopNavBar';

const YubiKeyAuth: React.FC = () => {
  const [otp, setOtp] = useState('');
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [currentUser, setCurrentUser] = useState<any>(null);
  const inputRef = useRef<HTMLInputElement>(null);
  
  const authenticateMutation = useAuthenticateDevice();
  const logoutMutation = useLogout();

  // Focus input on mount
  useEffect(() => {
    if (inputRef.current) {
      inputRef.current.focus();
    }
  }, []);

  // Check if user is authenticated
  useEffect(() => {
    if (currentUser) {
      setIsAuthenticated(true);
    } else {
      setIsAuthenticated(false);
    }
  }, [currentUser]);

  const handleOtpChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const value = e.target.value;
    
    // Accept Modhex characters (c, b, d, e, f, g, h, i, j, k, l, n, r, t, u, v)
    // and limit to 44 characters
    const modhexOnly = value.replace(/[^cbdhefghijklmnrtuv]/gi, '').slice(0, 44);
    
    setOtp(modhexOnly);
    
    // Don't auto-submit - let user complete the full OTP manually
    // YubiKey OTPs are generated in parts and we need the complete 44 characters
  };

  const handleAuthenticate = async () => {
    if (otp.length !== 44) return;

    // Debug logging to help identify the issue
    console.log('Original OTP length:', otp.length);
    console.log('Original OTP:', JSON.stringify(otp));
    console.log('Trimmed OTP length:', otp.trim().length);
    console.log('Trimmed OTP:', JSON.stringify(otp.trim()));

    try {
      const response = await authenticateMutation.mutateAsync({
        device_type: 'yubikey',
        auth_code: otp.trim(),
      });
      setOtp('');
      setCurrentUser(response.user);
    } catch (error) {
      console.error('Authentication failed:', error);
      setOtp('');
    }
  };

  const handleLogout = async () => {
    try {
      await logoutMutation.mutateAsync();
      setIsAuthenticated(false);
      setCurrentUser(null);
    } catch (error) {
      console.error('Logout failed:', error);
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && otp.length === 44) {
      handleAuthenticate();
    }
  };

  if (isAuthenticated && currentUser) {
    return (
      <>
        <TopNavBar />
        <div className="min-h-screen flex items-center justify-center bg-gray-50 font-sans">
          <div className="w-full max-w-md mx-auto">
            <div className="flex flex-col items-center mb-8">
              <div className="bg-blue-100 rounded-full p-3 mb-2">
                <svg className="w-10 h-10 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                </svg>
              </div>
            </div>
            <div className="bg-white rounded-xl shadow-lg px-8 py-10">
              <h1 className="text-2xl font-bold text-center mb-2 text-gray-900">Welcome, {currentUser.first_name}!</h1>
              <p className="text-gray-600 text-center mb-6">You are now authenticated with YubiApp</p>
              <div className="text-center text-gray-500 text-sm mb-6">
                <p><strong>Email:</strong> {currentUser.email}</p>
                <p><strong>Username:</strong> {currentUser.username}</p>
                <p><strong>Status:</strong> {currentUser.active ? 'Active' : 'Inactive'}</p>
              </div>
              <button
                onClick={handleLogout}
                disabled={logoutMutation.isPending}
                className="w-full bg-gray-100 hover:bg-gray-200 text-gray-800 font-semibold py-2 rounded-lg transition-colors duration-200"
              >
                {logoutMutation.isPending ? 'Signing out...' : 'Sign Out'}
              </button>
            </div>
          </div>
        </div>
      </>
    );
  }

  return (
    <>
      <TopNavBar />
      <div className="min-h-screen flex items-center justify-center bg-gray-50 font-sans">
        <div className="w-full max-w-md mx-auto">
          <div className="flex flex-col items-center mb-8">
            <div className="bg-blue-100 rounded-full p-3 mb-2">
              <svg className="w-10 h-10 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15.5 7.5a2 2 0 012 2m4 0a6 6 0 01-7.743 5.743L11 17H9v2H7v2H4a1 1 0 01-1-1v-2.586a1 1 0 01.293-.707l5.964-5.964A6 6 0 1121 9z" />
              </svg>
            </div>
          </div>
          <div className="bg-white rounded-xl shadow-lg px-8 py-10">
            <h1 className="text-2xl font-bold text-center mb-2 text-gray-900">YubiApp Authentication</h1>
            <p className="text-gray-600 text-center mb-6">Insert your YubiKey and tap it to authenticate</p>
            <form
              onSubmit={e => {
                e.preventDefault();
                handleAuthenticate();
              }}
            >
              <label htmlFor="otp" className="block text-sm font-medium text-gray-700 mb-2">
                YubiKey OTP
              </label>
              <input
                ref={inputRef}
                id="otp"
                type="text"
                value={otp}
                onChange={handleOtpChange}
                onKeyDown={handleKeyDown}
                placeholder="Tap your YubiKey here..."
                className="w-full px-5 py-3 rounded-lg border border-gray-300 bg-gray-50 text-lg font-mono tracking-wider focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent transition mb-2"
                maxLength={44}
                disabled={authenticateMutation.isPending}
                autoComplete="one-time-code"
                spellCheck={false}
                inputMode="text"
              />
              <div className="text-xs text-gray-500 mt-1 text-center mb-4">
                {otp.length}/44 characters
              </div>
              <button
                type="submit"
                disabled={otp.length !== 44 || authenticateMutation.isPending}
                className="w-full bg-blue-600 hover:bg-blue-700 text-white font-semibold py-3 rounded-lg text-lg transition-colors duration-200 mb-2"
              >
                {authenticateMutation.isPending ? 'Authenticating...' :
                  otp.length === 44 ? 'Authenticate' :
                  `Enter OTP (${otp.length}/44 characters)`}
              </button>
            </form>
            {authenticateMutation.isError && (
              <div className="bg-red-50 border border-red-200 rounded-lg p-3 mt-2 w-full text-center">
                <p className="text-red-600 text-sm">
                  Authentication failed. Please check your YubiKey and try again.
                </p>
              </div>
            )}
            <div className="text-xs text-gray-500 mt-6 text-center w-full">
              <p>Make sure your YubiKey is inserted and ready</p>
              <p>Tap the button on your YubiKey to generate an OTP</p>
            </div>
          </div>
        </div>
      </div>
    </>
  );
};

export default YubiKeyAuth; 