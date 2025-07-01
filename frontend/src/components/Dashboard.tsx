// @ts-nocheck
import React, { useEffect, useState } from 'react';
import TopNavBar from './TopNavBar';
import authFetch from '../services/authFetch';

const Dashboard: React.FC = () => {
  const [userStats, setUserStats] = useState<{ total: number; active: number } | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    const fetchUsers = async () => {
      setLoading(true);
      setError('');
      try {
        const res = await authFetch('/api/v1/users');
        const data = await res.json();
        if (!res.ok) throw new Error(data.error || 'Failed to fetch users');
        // Support both { users: [...] } and { items: [...] }
        const users = data.users || data.items || [];
        const total = users.length;
        const active = users.filter((u: any) => u.active).length;
        setUserStats({ total, active });
      } catch (err: any) {
        setError(err.message || 'Error fetching users');
      } finally {
        setLoading(false);
      }
    };
    fetchUsers();
  }, []);

  return (
    <>
      <TopNavBar />
      <div className="min-h-screen bg-gray-50 p-6">
        <div className="max-w-6xl mx-auto">
          <h1 className="text-2xl font-bold text-blue-700 mb-6">Dashboard</h1>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
            {/* Left column */}
            <div className="flex flex-col gap-8">
              <div className="bg-white rounded-lg shadow p-8 min-h-[220px] flex flex-col items-start justify-center">
                <span className="font-semibold text-2xl text-gray-700 mb-4">Users</span>
                {loading ? (
                  <span className="text-gray-400 text-lg">Loading...</span>
                ) : error ? (
                  <span className="text-red-500 text-lg">{error}</span>
                ) : userStats ? (
                  <>
                    <span className="text-4xl font-bold text-blue-700 mb-2">{userStats.active}</span>
                    <span className="text-gray-600 text-lg mb-1">active users</span>
                    <span className="text-gray-400 text-sm">Total users: {userStats.total}</span>
                  </>
                ) : null}
              </div>
              <div className="bg-white rounded-lg shadow p-8 min-h-[220px] flex flex-col items-start justify-center">
                <span className="font-semibold text-2xl text-gray-700 mb-4">Device Registration</span>
                <span className="text-gray-400 text-lg">(Device registration widget placeholder)</span>
              </div>
            </div>
            {/* Right column */}
            <div className="flex flex-col gap-8">
              <div className="bg-white rounded-lg shadow p-8 min-h-[220px] flex flex-col items-start justify-center">
                <span className="font-semibold text-2xl text-gray-700 mb-4">Actions</span>
                <span className="text-gray-400 text-lg">(Recent actions widget placeholder)</span>
              </div>
              <div className="bg-white rounded-lg shadow p-8 min-h-[220px] flex flex-col items-start justify-center">
                <span className="font-semibold text-2xl text-gray-700 mb-4">Message of the Day</span>
                <span className="text-gray-400 text-lg">(MOTD widget placeholder)</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </>
  );
};

export default Dashboard; 