<!--
  Modified by Jack de Haan, 2026 (meet fork of Timeful). See NOTICE.
  Replaces Timeful's "schejie sad" 404 with the jacksharks05.github.io 404
  design. The Galaxy WebGL backdrop there is replaced by meet's LightPillar
  backdrop (rendered behind every page in App.vue), so this just carries the
  matching veil + badge + title + subtitle.
-->
<template>
  <div class="notFound">
    <div class="notFound__veil"></div>

    <div class="notFound__content">
      <div class="notFound__badge">jack de haan</div>
      <h1 class="notFound__title">404</h1>
      <p class="notFound__subtitle">
        that page doesn't exist... just yet! head back
        <router-link :to="{ name: 'landing' }">home</router-link>.
      </p>
    </div>
  </div>
</template>

<script>
export default {
  name: "PageNotFound",

  metaInfo: {
    title: "404 - meet",
  },
}
</script>

<style scoped>
.notFound {
  position: relative;
  z-index: 1;
  min-height: 100vh;
  min-height: 100dvh;
  color: rgba(255, 255, 255, 0.92);
  overflow: hidden;
}

.notFound__veil {
  position: absolute;
  inset: 0;
  background:
    radial-gradient(
      800px 500px at 50% 40%,
      rgba(110, 170, 255, 0.1),
      transparent 60%
    ),
    radial-gradient(
      900px 600px at 55% 60%,
      rgba(180, 100, 255, 0.08),
      transparent 60%
    ),
    linear-gradient(to bottom, rgba(0, 0, 0, 0), rgba(0, 0, 0, 0.55));
  pointer-events: none;
}

.notFound__content {
  position: relative;
  z-index: 2;
  min-height: 100vh;
  min-height: 100dvh;
  display: grid;
  align-content: center;
  justify-items: center;
  padding: 48px 20px;
  text-align: center;
}

.notFound__badge {
  font-size: 12px;
  letter-spacing: 0.22em;
  text-transform: uppercase;
  padding: 10px 14px;
  border: 1px solid rgba(255, 255, 255, 0.12);
  border-radius: 999px;
  background: rgba(0, 0, 0, 0.25);
  backdrop-filter: blur(8px);
}

.notFound__title {
  margin: 18px 0 10px;
  font-size: clamp(40px, 6vw, 72px);
  line-height: 1.05;
  letter-spacing: -0.02em;
  font-weight: 700;
}

.notFound__subtitle {
  margin: 0;
  max-width: 62ch;
  font-size: 16px;
  line-height: 1.6;
  color: rgba(255, 255, 255, 0.78);
}

.notFound__subtitle a {
  color: rgba(160, 200, 255, 0.96);
  text-decoration: underline;
}
</style>
