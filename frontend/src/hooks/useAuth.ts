import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { apiService, type DeviceAuthRequest, type AuthResponse } from '../services/api';

export const useAuthenticateDevice = () => {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: (request: DeviceAuthRequest) => apiService.authenticateDevice(request),
    onSuccess: (data: AuthResponse) => {
      if (data.authenticated) {
        // Invalidate and refetch user data
        queryClient.invalidateQueries({ queryKey: ['currentUser'] });
      }
    },
  });
};

export const useCurrentUser = () => {
  return useQuery({
    queryKey: ['currentUser'],
    queryFn: () => apiService.getCurrentUser(),
    enabled: apiService.isAuthenticated(),
    staleTime: 5 * 60 * 1000, // 5 minutes
  });
};

export const useLogout = () => {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: () => {
      apiService.clearAuthToken();
      return Promise.resolve();
    },
    onSuccess: () => {
      // Clear all queries from cache
      queryClient.clear();
    },
  });
}; 