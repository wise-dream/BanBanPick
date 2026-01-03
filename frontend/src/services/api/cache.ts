/**
 * Кеширование API запросов
 */

interface CacheEntry<T> {
  data: T;
  timestamp: number;
  expiresAt: number;
}

interface CacheOptions {
  ttl?: number; // Time to live в миллисекундах
}

const DEFAULT_TTL = 5 * 60 * 1000; // 5 минут по умолчанию

class ApiCache {
  private cache: Map<string, CacheEntry<any>> = new Map();

  /**
   * Получить данные из кеша
   */
  get<T>(key: string): T | null {
    const entry = this.cache.get(key);
    
    if (!entry) {
      return null;
    }

    // Проверяем истечение
    if (Date.now() > entry.expiresAt) {
      this.cache.delete(key);
      return null;
    }

    return entry.data as T;
  }

  /**
   * Сохранить данные в кеш
   */
  set<T>(key: string, data: T, options: CacheOptions = {}): void {
    const ttl = options.ttl || DEFAULT_TTL;
    const now = Date.now();

    this.cache.set(key, {
      data,
      timestamp: now,
      expiresAt: now + ttl,
    });
  }

  /**
   * Удалить данные из кеша
   */
  delete(key: string): void {
    this.cache.delete(key);
  }

  /**
   * Очистить весь кеш
   */
  clear(): void {
    this.cache.clear();
  }

  /**
   * Инвалидировать кеш по паттерну ключа
   */
  invalidate(pattern: string | RegExp): void {
    const regex = typeof pattern === 'string' 
      ? new RegExp(pattern.replace('*', '.*'))
      : pattern;

    const keysToDelete: string[] = [];
    this.cache.forEach((_, key) => {
      if (regex.test(key)) {
        keysToDelete.push(key);
      }
    });

    keysToDelete.forEach(key => this.cache.delete(key));
  }

  /**
   * Очистить истекшие записи
   */
  cleanup(): void {
    const now = Date.now();
    const keysToDelete: string[] = [];

    this.cache.forEach((entry, key) => {
      if (now > entry.expiresAt) {
        keysToDelete.push(key);
      }
    });

    keysToDelete.forEach(key => this.cache.delete(key));
  }

  /**
   * Получить размер кеша
   */
  size(): number {
    return this.cache.size;
  }
}

// Глобальный экземпляр кеша
export const apiCache = new ApiCache();

/**
 * Генерация ключа кеша
 */
export function generateCacheKey(prefix: string, params: Record<string, any>): string {
  const sortedParams = Object.keys(params)
    .sort()
    .map(key => `${key}=${JSON.stringify(params[key])}`)
    .join('&');
  
  return `${prefix}:${sortedParams}`;
}

/**
 * Обертка для кеширования async функций
 */
export function withCache<T extends (...args: any[]) => Promise<any>>(
  fn: T,
  keyGenerator: (...args: Parameters<T>) => string,
  options: CacheOptions = {}
): T {
  return ((...args: Parameters<T>) => {
    const cacheKey = keyGenerator(...args);
    const cached = apiCache.get(cacheKey);

    if (cached !== null) {
      return Promise.resolve(cached);
    }

    return fn(...args).then(result => {
      apiCache.set(cacheKey, result, options);
      return result;
    });
  }) as T;
}

// Периодическая очистка истекших записей (каждые 5 минут)
if (typeof window !== 'undefined') {
  setInterval(() => {
    apiCache.cleanup();
  }, 5 * 60 * 1000);
}
