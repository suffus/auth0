// @ts-nocheck
import React from 'react';
import { Link, useLocation } from 'react-router-dom';

const TopNavBar: React.FC = () => {
  const location = useLocation();
  return (
    <nav className="bg-blue-700 text-white shadow">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex items-center justify-between h-14">
          {/* Left: Organization selector and app name */}
          <div className="flex items-center space-x-4">
            <div className="relative">
              <select className="bg-blue-700 text-white font-semibold rounded px-2 py-1 focus:outline-none">
                <option>Fanjango Limited</option>
                <option>Other Org (demo)</option>
              </select>
            </div>
            <span className="font-bold text-lg tracking-wide">Dashboard</span>
          </div>
          {/* Center/Right: Navigation links */}
          <div className="flex items-center space-x-6">
            <Link to="/dashboard" className={`hover:underline ${location.pathname === '/dashboard' ? 'font-bold underline' : ''}`}>Dashboard</Link>
            <Link to="/users" className="hover:underline text-white/80">Users</Link>
            <Link to="/actions" className="hover:underline text-white/80">Actions</Link>
            <Link to="/help" className="hover:underline text-white/80">Help</Link>
          </div>
          {/* Right: User info/avatar placeholder */}
          <div className="flex items-center space-x-3">
            <span className="text-sm text-white/80 hidden sm:inline">SF</span>
            <div className="w-8 h-8 rounded-full bg-blue-900 flex items-center justify-center font-bold">SF</div>
          </div>
        </div>
      </div>
    </nav>
  );
};

export default TopNavBar; 