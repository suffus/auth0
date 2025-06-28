import axios from 'axios';
import type { AxiosInstance, AxiosResponse } from 'axios';

// Types based on the OpenAPI specification
export interface User {
  id: string;
  email: string;
  username: string;
  first_name: string;
  last_name: string;
  active: boolean;
  created_at: string;
  updated_at: string;
  roles?: Role[];
}

export interface Role {
  id: string;
  name: string;
  description: string;
  created_at: string;
  updated_at: string;
  permissions?: Permission[];
}

export interface Resource {
  id: string;
  name: string;
  type: string;
  location: string;
  department: string;
  active: boolean;
  created_at: string;
  updated_at: string;
}

export interface Permission {
  id: string;
  resource: Resource;
  action: string;
  effect: string;
  created_at: string;
}

export interface Device {
  id: string;
  user: User;
  type: string;
  identifier: string;
  active: boolean;
  verified_at?: string;
  last_used_at?: string;
  created_at: string;
  updated_at: string;
}

export interface AuthResponse {
  authenticated: boolean;
  user: User;
}

export interface DeviceAuthRequest {
  device_type: string;
  auth_code: string;
  permission?: string;
}

export interface DeviceRegistrationRequest {
  target_user_id: string;
  device_identifier: string;
  device_type: 'yubikey' | 'totp' | 'sms' | 'email';
  notes?: string;
}

export interface ActionRequest {
  resource?: string;
  login?: string;
  [key: string]: any;
}

export interface ActionResponse {
  action: string;
  user_id: string;
  success: boolean;
  message: string;
}

class ApiService {
  private api: AxiosInstance;
  private authToken: string | null = null;

  constructor() {
    this.api = axios.create({
      baseURL: 'http://localhost:8080/api/v1',
      headers: {
        'Content-Type': 'application/json',
      },
    });

    // Add request interceptor to include auth token
    this.api.interceptors.request.use((config) => {
      if (this.authToken) {
        config.headers.Authorization = this.authToken;
      }
      return config;
    });

    // Add response interceptor for error handling
    this.api.interceptors.response.use(
      (response) => response,
      (error) => {
        if (error.response?.status === 401) {
          this.clearAuthToken();
        }
        return Promise.reject(error);
      }
    );
  }

  // Authentication methods
  async authenticateDevice(request: DeviceAuthRequest): Promise<AuthResponse> {
    const response: AxiosResponse<AuthResponse> = await this.api.post('/auth/device', request);
    if (response.data.authenticated) {
      this.authToken = `${request.device_type}:${request.auth_code}`;
    }
    return response.data;
  }

  setAuthToken(token: string) {
    this.authToken = token;
  }

  clearAuthToken() {
    this.authToken = null;
  }

  isAuthenticated(): boolean {
    return !!this.authToken;
  }

  // Device management methods
  async registerDevice(request: DeviceRegistrationRequest): Promise<Device> {
    const response: AxiosResponse<Device> = await this.api.post('/devices/register', request);
    return response.data;
  }

  async deregisterDevice(deviceId: string): Promise<void> {
    await this.api.post(`/devices/${deviceId}/deregister`);
  }

  async getDevices(userId?: string): Promise<Device[]> {
    const params = userId ? { user_id: userId } : {};
    const response: AxiosResponse<Device[]> = await this.api.get('/devices', { params });
    return response.data;
  }

  async getDevice(deviceId: string): Promise<Device> {
    const response: AxiosResponse<Device> = await this.api.get(`/devices/${deviceId}`);
    return response.data;
  }

  // Action methods
  async performAction(actionName: string, data: ActionRequest = {}): Promise<ActionResponse> {
    const response: AxiosResponse<ActionResponse> = await this.api.post(`/auth/action/${actionName}`, data);
    return response.data;
  }

  // User management methods
  async getCurrentUser(): Promise<User> {
    const response: AxiosResponse<User> = await this.api.get('/users/me');
    return response.data;
  }

  async getUsers(): Promise<User[]> {
    const response: AxiosResponse<User[]> = await this.api.get('/users');
    return response.data;
  }

  async getUser(userId: string): Promise<User> {
    const response: AxiosResponse<User> = await this.api.get(`/users/${userId}`);
    return response.data;
  }

  // Resource management methods
  async getResources(): Promise<Resource[]> {
    const response: AxiosResponse<Resource[]> = await this.api.get('/resources');
    return response.data;
  }

  async getResource(resourceId: string): Promise<Resource> {
    const response: AxiosResponse<Resource> = await this.api.get(`/resources/${resourceId}`);
    return response.data;
  }

  // Role management methods
  async getRoles(): Promise<Role[]> {
    const response: AxiosResponse<Role[]> = await this.api.get('/roles');
    return response.data;
  }

  async getRole(roleId: string): Promise<Role> {
    const response: AxiosResponse<Role> = await this.api.get(`/roles/${roleId}`);
    return response.data;
  }

  // Permission management methods
  async getPermissions(): Promise<Permission[]> {
    const response: AxiosResponse<Permission[]> = await this.api.get('/permissions');
    return response.data;
  }

  async getPermission(permissionId: string): Promise<Permission> {
    const response: AxiosResponse<Permission> = await this.api.get(`/permissions/${permissionId}`);
    return response.data;
  }
}

export const apiService = new ApiService();
export default apiService; 