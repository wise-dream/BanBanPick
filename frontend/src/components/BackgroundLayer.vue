<template>
  <div class="site-bg-layer">
    <div class="site-bg-inner"></div>
  </div>
</template>

<style scoped>
.site-bg-layer {
  position: fixed;
  inset: -50%;
  pointer-events: none;
  z-index: 0;
  overflow: hidden;
  transform: rotate(-30deg);
}

.site-bg-layer::before {
  content: "";
  position: absolute;
  inset: 0;
  background: rgba(0, 0, 0, 0.65);
}

.site-bg-inner {
  width: 200%;
  height: 200%;
  opacity: 0.5;
  background-image: url("/images/bg-text.svg");
  background-repeat: repeat;
  background-size: 400px 80px;
  animation: bgMove 180s linear infinite;
  will-change: transform;
}

@keyframes bgMove {
  from {
    transform: translate3d(0, 0, 0);
  }
  to {
    transform: translate3d(-400px, -1600px, 0);
  }
}
</style>
