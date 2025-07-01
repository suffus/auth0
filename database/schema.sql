-- YubiApp Database Schema
-- This file creates the complete database structure for the YubiApp authentication service

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create database if it doesn't exist (run this as superuser)
-- CREATE DATABASE yubiapp;

-- Connect to the database and run the following:

-- Users table
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    email VARCHAR(255) UNIQUE NOT NULL,
    username VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    first_name VARCHAR(255),
    last_name VARCHAR(255),
    active BOOLEAN DEFAULT TRUE
);

-- Roles table
CREATE TABLE roles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    name VARCHAR(255) UNIQUE NOT NULL,
    description TEXT
);

-- Resources table
CREATE TABLE resources (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    name VARCHAR(255) UNIQUE NOT NULL CHECK (name NOT LIKE '%:%'),
    type VARCHAR(50) NOT NULL CHECK (type IN ('server', 'service', 'database', 'application')),
    location VARCHAR(255),
    department VARCHAR(255),
    active BOOLEAN DEFAULT TRUE
);

-- Permissions table
CREATE TABLE permissions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    resource_id UUID NOT NULL REFERENCES resources(id) ON DELETE CASCADE,
    action VARCHAR(255) NOT NULL,
    effect VARCHAR(50) NOT NULL CHECK (effect IN ('allow', 'deny'))
);

-- Actions table
CREATE TABLE actions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    name VARCHAR(255) UNIQUE NOT NULL,
    activity_type VARCHAR(20) DEFAULT 'other' CHECK (activity_type IN ('user', 'system', 'automated', 'other')),
    required_permissions JSONB DEFAULT '[]'::jsonb,
    details JSONB DEFAULT '{}'::jsonb,
    active BOOLEAN DEFAULT true
);

-- Devices table
CREATE TABLE devices (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL CHECK (type IN ('yubikey', 'totp', 'sms', 'email')),
    identifier VARCHAR(255) NOT NULL,
    secret TEXT,
    last_used_at TIMESTAMP WITH TIME ZONE,
    verified_at TIMESTAMP WITH TIME ZONE,
    active BOOLEAN DEFAULT TRUE,
    properties JSONB
);

-- Authentication logs table
CREATE TABLE authentication_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    user_id UUID NOT NULL REFERENCES users(id),
    device_id UUID NOT NULL REFERENCES devices(id),
    action_id UUID REFERENCES actions(id) ON DELETE SET NULL,
    type VARCHAR(50) NOT NULL CHECK (type IN ('login', 'logout', 'refresh', 'mfa', 'action')),
    success BOOLEAN NOT NULL,
    ip_address VARCHAR(45),
    user_agent TEXT,
    details JSONB DEFAULT '{}'::jsonb
);

-- Device registrations table
CREATE TABLE device_registrations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    registrar_user_id UUID NOT NULL REFERENCES users(id),
    device_id UUID NOT NULL REFERENCES devices(id),
    target_user_id UUID REFERENCES users(id),
    
    action_type VARCHAR(20) NOT NULL CHECK (action_type IN ('register', 'deregister')),
    reason TEXT,
    ip_address VARCHAR(45),
    user_agent TEXT,
    notes TEXT
);

-- Locations table
CREATE TABLE locations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    name VARCHAR(255) UNIQUE NOT NULL,
    description TEXT,
    address TEXT,
    type VARCHAR(20) DEFAULT 'office' CHECK (type IN ('office', 'home', 'event', 'other')),
    active BOOLEAN DEFAULT TRUE
);

-- User statuses table
CREATE TABLE user_statuses (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    name VARCHAR(255) UNIQUE NOT NULL,
    description TEXT,
    type VARCHAR(30) DEFAULT 'working' CHECK (type IN ('working', 'break', 'leave', 'travel', 'other')),
    active BOOLEAN DEFAULT TRUE
);

-- User activity history table
CREATE TABLE user_activity_history (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    action_id UUID NOT NULL REFERENCES actions(id) ON DELETE CASCADE,
    from_datetime TIMESTAMP WITH TIME ZONE NOT NULL,
    to_datetime TIMESTAMP WITH TIME ZONE,
    location_id UUID REFERENCES locations(id) ON DELETE SET NULL,
    status_id UUID REFERENCES user_statuses(id) ON DELETE SET NULL,
    details JSONB DEFAULT '{}'::jsonb
);

-- Junction tables for many-to-many relationships

-- User-Role relationship
CREATE TABLE user_roles (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, role_id)
);

-- Role-Permission relationship
CREATE TABLE role_permissions (
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (role_id, permission_id)
);

-- Create indexes for better performance
CREATE INDEX idx_users_deleted_at ON users(deleted_at);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_active ON users(active);

CREATE INDEX idx_devices_deleted_at ON devices(deleted_at);
CREATE INDEX idx_devices_user_id ON devices(user_id);
CREATE INDEX idx_devices_type ON devices(type);
CREATE INDEX idx_devices_identifier ON devices(identifier);
CREATE INDEX idx_devices_active ON devices(active);

CREATE INDEX idx_authentication_logs_user_id ON authentication_logs(user_id);
CREATE INDEX idx_authentication_logs_device_id ON authentication_logs(device_id);
CREATE INDEX idx_authentication_logs_action_id ON authentication_logs(action_id);
CREATE INDEX idx_authentication_logs_created_at ON authentication_logs(created_at);
CREATE INDEX idx_authentication_logs_type ON authentication_logs(type);
CREATE INDEX idx_authentication_logs_success ON authentication_logs(success);

CREATE INDEX idx_locations_deleted_at ON locations(deleted_at);
CREATE INDEX idx_locations_name ON locations(name);
CREATE INDEX idx_locations_type ON locations(type);
CREATE INDEX idx_locations_active ON locations(active);

CREATE INDEX idx_user_statuses_deleted_at ON user_statuses(deleted_at);
CREATE INDEX idx_user_statuses_name ON user_statuses(name);
CREATE INDEX idx_user_statuses_type ON user_statuses(type);
CREATE INDEX idx_user_statuses_active ON user_statuses(active);

CREATE INDEX idx_user_activity_history_user_id ON user_activity_history(user_id);
CREATE INDEX idx_user_activity_history_action_id ON user_activity_history(action_id);
CREATE INDEX idx_user_activity_history_from_datetime ON user_activity_history(from_datetime);
CREATE INDEX idx_user_activity_history_to_datetime ON user_activity_history(to_datetime);
CREATE INDEX idx_user_activity_history_location_id ON user_activity_history(location_id);
CREATE INDEX idx_user_activity_history_status_id ON user_activity_history(status_id);
CREATE INDEX idx_user_activity_history_user_from_datetime ON user_activity_history(user_id, from_datetime);

CREATE INDEX idx_actions_name ON actions(name);

CREATE INDEX idx_user_roles_user_id ON user_roles(user_id);
CREATE INDEX idx_user_roles_role_id ON user_roles(role_id);

CREATE INDEX idx_role_permissions_role_id ON role_permissions(role_id);
CREATE INDEX idx_role_permissions_permission_id ON role_permissions(permission_id);

-- Create updated_at trigger function
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create triggers for updated_at
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_roles_updated_at BEFORE UPDATE ON roles FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_permissions_updated_at BEFORE UPDATE ON permissions FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_actions_updated_at BEFORE UPDATE ON actions FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_devices_updated_at BEFORE UPDATE ON devices FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_locations_updated_at BEFORE UPDATE ON locations FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_user_statuses_updated_at BEFORE UPDATE ON user_statuses FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_user_activity_history_updated_at BEFORE UPDATE ON user_activity_history FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Insert some default data
-- Default admin role
INSERT INTO roles (id, name, description) VALUES 
    (uuid_generate_v4(), 'admin', 'Administrator with full access'),
    (uuid_generate_v4(), 'user', 'Standard user');

-- Default resources
INSERT INTO resources (id, name, type, location, department) VALUES 
    (uuid_generate_v4(), 'user', 'application', 'internal', 'IT'),
    (uuid_generate_v4(), 'device', 'application', 'internal', 'IT'),
    (uuid_generate_v4(), 'admin', 'application', 'internal', 'IT'),
    (uuid_generate_v4(), 'yubiapp', 'application', 'internal', 'IT');

-- Default permissions
INSERT INTO permissions (id, resource_id, action, effect) 
SELECT uuid_generate_v4(), r.id, p.action, p.effect
FROM resources r
CROSS JOIN (VALUES ('read', 'allow'), ('write', 'allow')) AS p(action, effect)
WHERE r.name IN ('user', 'device', 'admin');

-- Add yubiapp-specific permissions
INSERT INTO permissions (id, resource_id, action, effect)
SELECT uuid_generate_v4(), r.id, p.action, p.effect
FROM resources r
CROSS JOIN (VALUES ('register-other', 'allow'), ('deregister-other', 'allow')) AS p(action, effect)
WHERE r.name = 'yubiapp';

-- Default actions
INSERT INTO actions (id, name, activity_type, required_permissions, details) VALUES 
    (uuid_generate_v4(), 'ssh-login', 'user', '["ssh:login"]', '{}'),
    (uuid_generate_v4(), 'app-install', 'system', '["app:install"]', '{}'),
    (uuid_generate_v4(), 'app-uninstall', 'system', '["app:uninstall"]', '{}'),
    (uuid_generate_v4(), 'permission-grant', 'system', '["permission:grant"]', '{}'),
    (uuid_generate_v4(), 'permission-revoke', 'system', '["permission:revoke"]', '{}'),
    (uuid_generate_v4(), 'user-signin', 'user', '[]', '{}'),
    (uuid_generate_v4(), 'user-signout', 'user', '[]', '{}'),
    (uuid_generate_v4(), 'status-change', 'system', '[]', '{}'),
    (uuid_generate_v4(), 'location-change', 'system', '[]', '{}'),
    (uuid_generate_v4(), 'break-start', 'system', '[]', '{}'),
    (uuid_generate_v4(), 'break-end', 'system', '[]', '{}'),
    (uuid_generate_v4(), 'work-start', 'system', '[]', '{}'),
    (uuid_generate_v4(), 'work-end', 'system', '[]', '{}'),
    (uuid_generate_v4(), 'meeting-start', 'system', '[]', '{}'),
    (uuid_generate_v4(), 'meeting-end', 'system', '[]', '{}');

-- Default user statuses
INSERT INTO user_statuses (id, name, description, type, active) VALUES 
    (uuid_generate_v4(), 'Signed In', 'User is signed in and available', 'working', true),
    (uuid_generate_v4(), 'On Break', 'User is taking a break', 'break', true),
    (uuid_generate_v4(), 'In Meeting', 'User is in a meeting', 'working', true),
    (uuid_generate_v4(), 'Out of Office', 'User is out of office', 'leave', true),
    (uuid_generate_v4(), 'Traveling', 'User is traveling for work', 'travel', true),
    (uuid_generate_v4(), 'Lunch Break', 'User is on lunch break', 'break', true),
    (uuid_generate_v4(), 'Training', 'User is in training', 'working', true),
    (uuid_generate_v4(), 'Sick Leave', 'User is on sick leave', 'leave', true),
    (uuid_generate_v4(), 'Vacation', 'User is on vacation', 'leave', true);

-- Default locations
INSERT INTO locations (id, name, description, address, type, active) VALUES 
    (uuid_generate_v4(), 'Main Office', 'Primary office location', '123 Main Street, City, State 12345', 'office', true),
    (uuid_generate_v4(), 'Home Office', 'Working from home', 'Various home locations', 'home', true),
    (uuid_generate_v4(), 'Client Site', 'Working at client location', 'Client office locations', 'other', true),
    (uuid_generate_v4(), 'Conference Center', 'Conference and event location', '456 Conference Ave, City, State 12345', 'event', true),
    (uuid_generate_v4(), 'Airport Lounge', 'Working from airport', 'Various airport lounges', 'other', true),
    (uuid_generate_v4(), 'Hotel Room', 'Working from hotel', 'Various hotel locations', 'other', true),
    (uuid_generate_v4(), 'Coffee Shop', 'Working from coffee shop', 'Various coffee shops', 'other', true),
    (uuid_generate_v4(), 'Co-working Space', 'Shared office space', '789 Co-working Blvd, City, State 12345', 'office', true);

-- Assign permissions to roles
-- Admin gets all permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id 
FROM roles r, permissions p 
WHERE r.name = 'admin';

-- User gets basic permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id 
FROM roles r, permissions p, resources res
WHERE r.name = 'user' 
AND p.resource_id = res.id
AND res.name IN ('user', 'device') 
AND p.action = 'read'; 