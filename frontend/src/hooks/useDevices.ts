import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { apiService, type DeviceRegistrationRequest } from '../services/api';

export const useDevices = (userId?: string) => {
  return useQuery({
    queryKey: ['devices', userId],
    queryFn: () => apiService.getDevices(userId),
    enabled: apiService.isAuthenticated(),
  });
};

export const useDevice = (deviceId: string) => {
  return useQuery({
    queryKey: ['device', deviceId],
    queryFn: () => apiService.getDevice(deviceId),
    enabled: !!deviceId && apiService.isAuthenticated(),
  });
};

export const useRegisterDevice = () => {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: (request: DeviceRegistrationRequest) => apiService.registerDevice(request),
    onSuccess: () => {
      // Invalidate devices queries
      queryClient.invalidateQueries({ queryKey: ['devices'] });
    },
  });
};

export const useDeregisterDevice = () => {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: (deviceId: string) => apiService.deregisterDevice(deviceId),
    onSuccess: () => {
      // Invalidate devices queries
      queryClient.invalidateQueries({ queryKey: ['devices'] });
    },
  });
}; 