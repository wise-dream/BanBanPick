import { createApp } from 'vue';
import { createI18n } from 'vue-i18n';
import { router } from './router';
import { pinia } from './store';
import { useAuthStore } from './store/auth';
import App from './App.vue';
import './style.css';

// PrimeVue setup
import PrimeVue from 'primevue/config';
import Aura from '@primevue/themes/aura';
import 'primeicons/primeicons.css';

// i18n setup
import en from './locales/en.json';
import ru from './locales/ru.json';

const i18n = createI18n({
  legacy: false,
  locale: localStorage.getItem('locale') || 'en',
  fallbackLocale: 'en',
  messages: {
    en,
    ru
  }
});

const app = createApp(App);

app.use(pinia);
app.use(router);
app.use(i18n);
app.use(PrimeVue, {
  theme: {
    preset: Aura,
    options: {
      darkModeSelector: '.dark-mode',
      cssLayer: false
    }
  }
});

// Инициализация auth store
const authStore = useAuthStore();
authStore.initAuth();

app.mount('#app');
