<script setup lang="ts">
interface Props {
  currentStep: number;
  totalSteps?: number;
}

const props = withDefaults(defineProps<Props>(), {
  totalSteps: 3
});

const steps = Array.from({ length: props.totalSteps }, (_, i) => i + 1);
</script>

<template>
  <div class="step-indicator">
    <div
      v-for="step in steps"
      :key="step"
      class="step-item"
      :class="{ active: step === currentStep, completed: step < currentStep }"
    >
      <div class="step-number">{{ step }}</div>
      <div class="step-label">Step {{ step }}</div>
    </div>
  </div>
</template>

<style scoped>
.step-indicator {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 2rem;
  margin: 2rem 0;
}

.step-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.5rem;
  position: relative;
  opacity: 0.5;
  transition: opacity 0.3s;
}

.step-item.active {
  opacity: 1;
}

.step-item.completed {
  opacity: 0.8;
}

.step-number {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background: rgba(0, 0, 0, 0.9);
  border: 2px solid rgba(255, 255, 255, 0.3);
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: bold;
  color: white;
  transition: all 0.3s;
}

.step-item.active .step-number {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-color: transparent;
  transform: scale(1.1);
}

.step-item.completed .step-number {
  background: rgba(76, 175, 80, 0.8);
  border-color: rgba(76, 175, 80, 1);
}

.step-label {
  font-size: 0.85rem;
  color: rgba(255, 255, 255, 0.7);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.step-item.active .step-label {
  color: white;
  font-weight: 600;
}

@media (max-width: 768px) {
  .step-indicator {
    gap: 1rem;
  }

  .step-number {
    width: 32px;
    height: 32px;
    font-size: 0.9rem;
  }

  .step-label {
    font-size: 0.75rem;
  }
}
</style>
