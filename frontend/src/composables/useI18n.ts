import { computed } from 'vue';
import { useI18n as useVueI18n } from 'vue-i18n';

export interface Locale {
  code: string;
  name: string;
}

export const availableLocales: Locale[] = [
  { code: 'en', name: 'English' },
  { code: 'ru', name: 'Русский' }
];

export function useI18n() {
  const { locale, t } = useVueI18n();
  
  // Безопасное получение текущей локали с fallback на 'en'
  const currentLocale = computed(() => {
    try {
      return locale.value || localStorage.getItem('locale') || 'en';
    } catch {
      return 'en';
    }
  });

  const setLocale = (newLocale: string) => {
    if (availableLocales.some(l => l.code === newLocale)) {
      try {
        locale.value = newLocale;
        localStorage.setItem('locale', newLocale);
      } catch (error) {
        console.error('Error setting locale:', error);
      }
    }
  };

  return {
    currentLocale,
    availableLocales,
    setLocale,
    t
  };
}
