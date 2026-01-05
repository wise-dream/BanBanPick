<script setup lang="ts">
import { computed } from 'vue';
import { Globe } from 'lucide-vue-next';
import { useI18n } from '../../composables/useI18n';
import Select from 'primevue/select';

const { currentLocale, availableLocales, setLocale, t } = useI18n();

// Безопасное получение текущей локали с fallback
const safeCurrentLocale = computed(() => {
  return currentLocale.value || 'en';
});

// Преобразуем локали в формат для dropdown
interface LocaleOption {
  label: string;
  value: string;
  code: string;
}

const localeOptions = computed<LocaleOption[]>(() => {
  return availableLocales.map((locale) => ({
    label: locale.name,
    value: locale.code,
    code: locale.code
  }));
});

const selectedLocale = computed({
  get: () => localeOptions.value.find((opt: LocaleOption) => opt.value === safeCurrentLocale.value) || localeOptions.value[0],
  set: (newLocale: LocaleOption | null) => {
    if (newLocale && newLocale.value) {
      setLocale(newLocale.value);
    }
  }
});
</script>

<template>
  <div class="language-selector">
    <Select
      v-model="selectedLocale"
      :options="localeOptions"
      optionLabel="label"
      :placeholder="t('languageSelector.placeholder')"
      class="language-select-pv"
    >
      <template #value="slotProps">
        <div v-if="slotProps.value" class="language-selected">
          <Globe :size="18" />
          <span>{{ slotProps.value.value.toUpperCase() }}</span>
        </div>
        <span v-else>{{ slotProps.placeholder }}</span>
      </template>
    </Select>
  </div>
</template>

<style scoped>
.language-selector {
  display: flex;
  align-items: center;
}

.language-select-pv {
  min-width: 100px;
  height: 36px;
}

:deep(.language-select-pv) {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  background: rgba(0, 0, 0, 0.9);
  border: 1px solid rgba(255, 255, 255, 0.2);
  border-radius: 4px;
  padding: 0.5rem 0.75rem;
  font-size: 0.85rem;
}

.language-selected {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  color: white;
  font-weight: 500;
}
</style>
