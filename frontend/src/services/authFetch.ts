// @ts-nocheck

/**
 * authFetch: fetch wrapper that adds Authorization header and refreshes access token if needed.
 * Usage: authFetch(url, options)
 */
export default async function authFetch(url: string, options: any = {}) {
  let accessToken = localStorage.getItem('access_token');
  let refreshToken = localStorage.getItem('refresh_token');
  let sessionId = localStorage.getItem('session_id');

  // Add Authorization header
  options.headers = options.headers || {};
  if (accessToken) {
    options.headers['Authorization'] = `Bearer ${accessToken}`;
  }

  let res = await fetch(url, options);

  // If unauthorized, try to refresh the token
  if (res.status === 401 && refreshToken && sessionId) {
    // Try to refresh
    const refreshRes = await fetch(`/api/v1/auth/session/refresh/${sessionId}`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ refresh_token: refreshToken }),
    });
    const refreshData = await refreshRes.json();
    if (refreshRes.ok && refreshData.access_token) {
      // Save new tokens
      localStorage.setItem('access_token', refreshData.access_token);
      localStorage.setItem('refresh_token', refreshData.refresh_token);
      // Update header and retry original request
      options.headers['Authorization'] = `Bearer ${refreshData.access_token}`;
      res = await fetch(url, options);
    } else {
      // Refresh failed, clear tokens
      localStorage.removeItem('access_token');
      localStorage.removeItem('refresh_token');
      localStorage.removeItem('session_id');
      throw new Error('Session expired. Please log in again.');
    }
  }

  return res;
} 