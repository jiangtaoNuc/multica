import { QueryClient } from "@tanstack/react-query";
import { NetworkError } from "./api/client";

export function createQueryClient(): QueryClient {
  return new QueryClient({
    defaultOptions: {
      queries: {
        staleTime: Infinity,
        gcTime: 10 * 60 * 1000, // 10 minutes
        refetchOnWindowFocus: false,
        refetchOnReconnect: true,
        retry: (failureCount, error) => {
          if (error instanceof NetworkError) return failureCount < 3;
          return failureCount < 1;
        },
        retryDelay: (attemptIndex) =>
          Math.min(1000 * Math.pow(2, attemptIndex), 15000),
      },
      mutations: {
        retry: false,
      },
    },
  });
}
