<script setup lang="ts">
import { useRouter } from 'vue-router';

const router = useRouter();
</script>

<template>
  <div class="not-found-page">
    <div class="not-found-content">
      <h1>404</h1>
      <p>Page not found</p>
      <button @click="router.push('/')" class="btn-home">
        Go to Home
      </button>
    </div>
  </div>
</template>

<style scoped>
.not-found-page {
  min-height: calc(100vh - 200px);
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
  z-index: 1;
}

.not-found-content {
  text-align: center;
  color: white;
  position: relative;
  z-index: 1;
}

.not-found-content h1 {
  font-size: 6rem;
  margin: 0;
  font-weight: 700;
}

.not-found-content p {
  font-size: 1.5rem;
  margin: 1rem 0 2rem;
}

.btn-home {
  background: white;
  color: #667eea;
  border: none;
  padding: 1rem 2rem;
  border-radius: 8px;
  font-size: 1rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s;
}

.btn-home:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.2);
}
</style>
