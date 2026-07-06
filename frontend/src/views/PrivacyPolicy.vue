<!--
  Modified by Jack de Haan, 2026 (meet fork of Timeful). See NOTICE.
  Replaces Timeful's embedded Google-Doc privacy policy (their document, and
  inaccurate for this fork) with a short, honest notice for meet.jackdehaan.com.
-->
<template>
  <div class="jdh-privacy">
    <div class="jdh-card jdh-privacy__card">
      <h1 class="jdh-privacy__title">Privacy</h1>
      <p class="jdh-privacy__updated">meet.jackdehaan.com</p>

      <p>
        meet is a private, self-hosted availability poll. It is a fork of the
        open-source
        <a
          href="https://github.com/schej-it/timeful.app"
          target="_blank"
          rel="noopener"
          >Timeful</a
        >
        app, with the marketing site, payments, ads, and third-party analytics
        removed.
      </p>

      <h2>What is collected</h2>
      <p>
        When you respond to a poll, meet stores the name you enter and the time
        slots you mark. If a specific poll asks for it, it also stores the email
        address you provide. That is all.
      </p>

      <h2>Where it goes</h2>
      <p>
        Responses are stored on a private, self-hosted server and are visible to
        the person who created the poll. Nothing is sold or shared with anyone
        else, and there are no advertising or analytics trackers on this site.
      </p>

      <h2>Your data</h2>
      <p>
        You can update or remove your own response from a poll at any time. To
        have a poll and its responses deleted entirely, contact the poll's
        creator (meet [at] jackdehaan [dot] com).
      </p>

      <div class="jdh-privacy__rule"></div>
      <a
        class="jdh-privacy__src"
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
export default {
  name: "PrivacyPolicy",

  metaInfo: {
    title: "Privacy - meet",
  },
}
</script>

<style scoped>
.jdh-privacy {
  position: relative;
  z-index: 1;
  min-height: 100vh;
  display: flex;
  align-items: flex-start;
  justify-content: center;
  padding: 40px 24px;
}
.jdh-privacy__card {
  max-width: 720px;
  width: 100%;
  padding: 32px 28px;
  color: var(--jdh-muted);
  line-height: 1.65;
}
.jdh-privacy__title {
  font-size: 34px;
  font-weight: 700;
  color: #fff;
  margin: 0 0 4px;
}
.jdh-privacy__updated {
  font-size: 13px;
  color: rgba(255, 255, 255, 0.5);
  margin: 0 0 20px;
}
.jdh-privacy__card h2 {
  color: #fff;
  font-size: 18px;
  font-weight: 600;
  margin: 20px 0 6px;
}
.jdh-privacy__card p {
  margin: 0 0 12px;
}
.jdh-privacy__card a {
  color: var(--jdh-link);
  text-decoration: none;
}
.jdh-privacy__card a:hover {
  text-decoration: underline;
}
.jdh-privacy__rule {
  height: 1px;
  margin: 22px auto 16px;
  width: 100%;
  background: linear-gradient(
    90deg,
    transparent,
    var(--jdh-outline),
    transparent
  );
}
.jdh-privacy__src {
  font-size: 12px;
  letter-spacing: 0.3px;
}
</style>
