<!--
  Modified by Jack de Haan, 2026 (meet fork of Timeful). See NOTICE.
  Replaces Timeful's marketing homepage with a minimal dark landing for
  meet.jackdehaan.com. The LightPillar backdrop (App.vue) shows through.
  The Google Calendar booking page is embedded inline (gv=true embed view).
-->
<template>
  <div class="jdh-landing">
    <div class="jdh-card jdh-landing__card">
      <h1 class="jdh-landing__title">meet with jack</h1>
      <p class="jdh-landing__lede">exciting times ahead!</p>

      <div class="jdh-landing__booking">
        <iframe
          :src="bookingUrl"
          class="jdh-landing__iframe"
          title="Book a time with Jack"
          loading="lazy"
        ></iframe>
      </div>

      <p class="jdh-landing__sub">
        Have a group poll link? Open it to mark when you're free.
      </p>
      <div class="jdh-landing__rule"></div>
      <a
        class="jdh-landing__src"
        href="https://github.com/JackSharks05/meet"
        target="_blank"
        rel="noopener"
      >
        Open source (AGPL-3.0) — view the source
      </a>
    </div>
  </div>
</template>

<script>
// Google Calendar appointment-scheduling link. Override at build time with
// VUE_APP_BOOKING_URL; falls back to Jack's booking page. gv=true renders the
// embeddable view, so it can be shown inline in an iframe.
const DEFAULT_BOOKING_URL =
  "https://calendar.google.com/calendar/appointments/schedules/AcZssZ1O7q4a8cqz90Lla7qIxwwwbKldW6GYz4MslSQe3KMLJWag5XJ2J4Gxl5O2_A7SBuTBc5EJnH6W?gv=true"

export default {
  name: "Landing",
  computed: {
    bookingUrl() {
      return process.env.VUE_APP_BOOKING_URL || DEFAULT_BOOKING_URL
    },
  },
}
</script>

<style scoped>
.jdh-landing {
  position: relative;
  z-index: 1;
  min-height: 100vh;
  display: flex;
  align-items: flex-start;
  justify-content: center;
  padding: 40px 24px;
}
.jdh-landing__card {
  max-width: 1080px;
  width: 100%;
  padding: 32px 28px;
  text-align: center;
}
.jdh-landing__title {
  font-size: 44px;
  font-weight: 700;
  letter-spacing: 1px;
  margin: 0 0 8px;
  color: #fff;
}
.jdh-landing__lede {
  font-size: 18px;
  color: var(--jdh-muted);
  line-height: 1.5;
  margin: 0 0 18px;
}
.jdh-landing__booking {
  background: #fff;
  border-radius: 14px;
  overflow: hidden;
  border: 1px solid var(--jdh-outline);
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.35);
}
.jdh-landing__iframe {
  display: block;
  width: 100%;
  height: clamp(520px, 70vh, 760px);
  border: 0;
}
.jdh-landing__sub {
  font-size: 14px;
  color: rgba(255, 255, 255, 0.55);
  line-height: 1.65;
  margin: 16px 0 0;
}
.jdh-landing__rule {
  height: 1px;
  margin: 18px auto 16px;
  width: 70%;
  background: linear-gradient(
    90deg,
    transparent,
    var(--jdh-outline),
    transparent
  );
}
.jdh-landing__src {
  font-size: 12px;
  letter-spacing: 0.3px;
  color: var(--jdh-link);
  text-decoration: none;
}
.jdh-landing__src:hover {
  text-decoration: underline;
}

@media (max-width: 600px) {
  .jdh-landing__title {
    font-size: 36px;
  }
}
</style>
