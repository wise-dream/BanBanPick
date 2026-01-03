import { createApp } from 'vue';
import { createI18n } from 'vue-i18n';
import { router } from './router';
import { pinia } from './store';
import { useAuthStore } from './store/auth';
import App from './App.vue';
import './style.css';

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

// Инициализация auth store
const authStore = useAuthStore();
authStore.initAuth();

app.mount('#app');
