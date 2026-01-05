<script setup lang="ts">
import { useRouter } from 'vue-router';
import { useI18n } from '../../composables/useI18n';

const { t } = useI18n();

const router = useRouter();
</script>

<template>
  <footer class="app-footer">
    <div class="footer-container">
      <div class="footer-logo" @click="router.push('/')">
        <h2>MapBan</h2>
      </div>
      <div class="footer-bottom">
        <p>&copy; {{ new Date().getFullYear() }} MapBan – {{ t('footer.copyright') }}</p>
      </div>
    </div>
  </footer>
</template>

<style scoped>
.app-footer {
  background: rgba(0, 0, 0, 0.9);
  border-top: 1px solid rgba(255, 255, 255, 0.1);
  padding: 2rem;
  margin-top: auto;
  color: white;
}

.footer-container {
  max-width: 1400px;
  margin: 0 auto;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 1rem;
}

.footer-logo {
  cursor: pointer;
}

.footer-logo h2 {
  margin: 0;
  font-size: 1.5rem;
  font-weight: bold;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.footer-bottom {
  text-align: center;
}

.footer-bottom p {
  color: rgba(255, 255, 255, 0.5);
  margin: 0;
  font-size: 0.85rem;
}
</style>
