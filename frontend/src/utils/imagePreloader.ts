/**
 * Утилиты для предзагрузки изображений
 */

interface PreloadOptions {
  priority?: 'high' | 'low' | 'auto';
  fetchPriority?: 'high' | 'low' | 'auto';
}

/**
 * Предзагружает изображение
 */
export function preloadImage(src: string, options: PreloadOptions = {}): Promise<void> {
  return new Promise((resolve, reject) => {
    const link = document.createElement('link');
    link.rel = 'preload';
    link.as = 'image';
    link.href = src;
    
    if (options.priority) {
      link.setAttribute('importance', options.priority);
    }
    
    if (options.fetchPriority) {
      link.setAttribute('fetchpriority', options.fetchPriority);
    }

    link.onload = () => resolve();
    link.onerror = () => reject(new Error(`Failed to preload image: ${src}`));
    
    document.head.appendChild(link);
  });
}

/**
 * Предзагружает несколько изображений
 */
export async function preloadImages(
  sources: string[],
  options: PreloadOptions = {}
): Promise<void[]> {
  return Promise.all(sources.map(src => preloadImage(src, options)));
}

/**
 * Lazy loading для изображений
 */
export function setupLazyLoading(selector: string = 'img[data-src]'): void {
  if ('IntersectionObserver' in window) {
    const imageObserver = new IntersectionObserver((entries, observer) => {
      entries.forEach(entry => {
        if (entry.isIntersecting) {
          const img = entry.target as HTMLImageElement;
          const src = img.getAttribute('data-src');
          if (src) {
            img.src = src;
            img.removeAttribute('data-src');
            observer.unobserve(img);
          }
        }
      });
    });

    document.querySelectorAll(selector).forEach(img => {
      imageObserver.observe(img);
    });
  } else {
    // Fallback для браузеров без IntersectionObserver
    document.querySelectorAll(selector).forEach((img: Element) => {
      const imgElement = img as HTMLImageElement;
      const src = imgElement.getAttribute('data-src');
      if (src) {
        imgElement.src = src;
        imgElement.removeAttribute('data-src');
      }
    });
  }
}

/**
 * Предзагружает критичные изображения карт
 */
export function preloadCriticalMapImages(mapNames: string[]): void {
  const criticalImages = mapNames.slice(0, 3).map(name => 
    `/images/maps/${name.toLowerCase()}.jpg`
  );
  
  preloadImages(criticalImages, { priority: 'high', fetchPriority: 'high' }).catch(err => {
    console.warn('Failed to preload some map images:', err);
  });
}
