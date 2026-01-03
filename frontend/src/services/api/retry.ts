import type { AxiosError } from 'axios';
import type { ApiError } from './types';
import { canRetry, isNetworkError } from './errorHandler';

/**
 * Retry configuration
 */
export interface RetryConfig {
  maxRetries: number;
  retryDelay: number;
  retryCondition?: (error: AxiosError) => boolean;
}

const defaultConfig: RetryConfig = {
  maxRetries: 3,
  retryDelay: 1000, // 1 second
};

/**
 * Retry a failed request
 */
export async function retryRequest<T>(
  requestFn: () => Promise<T>,
  config: Partial<RetryConfig> = {}
): Promise<T> {
  const retryConfig = { ...defaultConfig, ...config };
  let lastError: AxiosError<ApiError> | null = null;

  for (let attempt = 0; attempt <= retryConfig.maxRetries; attempt++) {
    try {
      return await requestFn();
    } catch (error) {
      lastError = error as AxiosError<ApiError>;

      // Check if we should retry
      if (attempt < retryConfig.maxRetries) {
        const shouldRetry = retryConfig.retryCondition
          ? retryConfig.retryCondition(lastError)
          : canRetry(lastError.response?.data || lastError);

        if (shouldRetry) {
          // Wait before retrying (exponential backoff)
          const delay = retryConfig.retryDelay * Math.pow(2, attempt);
          await new Promise(resolve => setTimeout(resolve, delay));
          continue;
        }
      }

      throw error;
    }
  }

  throw lastError;
}

/**
 * Creates a retry wrapper for API calls
 */
export function withRetry<T extends (...args: any[]) => Promise<any>>(
  fn: T,
  config?: Partial<RetryConfig>
): T {
  return ((...args: Parameters<T>) => {
    return retryRequest(() => fn(...args), config);
  }) as T;
}
