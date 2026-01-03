<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue';
import { Globe } from 'lucide-vue-next';
import { useI18n } from '../../composables/useI18n';

const { currentLocale, availableLocales, setLocale } = useI18n();
const isOpen = ref(false);
const dropdownRef = ref<HTMLElement | null>(null);

// Безопасное получение текущей локали с fallback
const safeCurrentLocale = computed(() => {
  return currentLocale.value || 'en';
});

const currentLocaleCode = computed(() => {
  return safeCurrentLocale.value.toUpperCase();
});

const changeLanguage = (localeCode: string) => {
  setLocale(localeCode);
  isOpen.value = false;
};

const toggleDropdown = () => {
  isOpen.value = !isOpen.value;
};

const closeDropdown = (event: MouseEvent) => {
  if (dropdownRef.value && !dropdownRef.value.contains(event.target as Node)) {
    isOpen.value = false;
  }
};

onMounted(() => {
  document.addEventListener('click', closeDropdown);
});

onUnmounted(() => {
  document.removeEventListener('click', closeDropdown);
});
</script>

<template>
  <div class="language-selector" ref="dropdownRef">
    <button
      class="language-button"
      @click="toggleDropdown"
      :aria-expanded="isOpen"
      aria-label="Select language"
    >
      <Globe :size="18" />
      <span class="language-code">{{ currentLocaleCode }}</span>
    </button>
    <div v-if="isOpen" class="language-dropdown">
      <button
        v-for="locale in availableLocales"
        :key="locale.code"
        class="language-option"
        :class="{ active: safeCurrentLocale === locale.code }"
        @click="changeLanguage(locale.code)"
      >
        {{ locale.name }}
      </button>
    </div>
  </div>
</template>

<style scoped>
.language-selector {
  position: relative;
  display: flex;
  align-items: center;
}

.language-button {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  background: rgba(0, 0, 0, 0.9);
  border: 1px solid rgba(255, 255, 255, 0.2);
  border-radius: 4px;
  color: white;
  padding: 0.5rem 0.75rem;
  font-size: 0.85rem;
  cursor: pointer;
  transition: all 0.2s;
}

.language-button:hover {
  background: rgba(255, 255, 255, 0.2);
  border-color: rgba(255, 255, 255, 0.3);
}

.language-code {
  font-weight: 500;
}

.language-dropdown {
  position: absolute;
  top: calc(100% + 0.5rem);
  right: 0;
  background: rgba(0, 0, 0, 0.95);
  backdrop-filter: blur(10px);
  border: 1px solid rgba(255, 255, 255, 0.2);
  border-radius: 4px;
  min-width: 150px;
  overflow: hidden;
  z-index: 1000;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
}

.language-option {
  display: block;
  width: 100%;
  text-align: left;
  background: transparent;
  border: none;
  color: white;
  padding: 0.75rem 1rem;
  font-size: 0.9rem;
  cursor: pointer;
  transition: background 0.2s;
}

.language-option:hover {
  background: rgba(0, 0, 0, 0.9);
}

.language-option.active {
  background: rgba(102, 126, 234, 0.2);
  color: #667eea;
  font-weight: 500;
}
</style>
