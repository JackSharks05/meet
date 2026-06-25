// Modified by Jack de Haan, 2026 (meet fork of Timeful). See NOTICE.
// Third-party analytics (PostHog + Google Tag Manager) removed in this fork.
import Vue from "vue"
import VueWorker from "vue-worker"
import App from "./App.vue"
import router from "./router"
import store from "./store"
import vuetify from "./plugins/vuetify"
import posthogPlugin from "./plugins/posthog"
import VueMeta from "vue-meta"
import "./index.css"

// Analytics: $posthog is a no-op stub (see plugins/posthog.js); no data is sent.
Vue.use(posthogPlugin)

// Site Metadata
Vue.use(VueMeta)

// Workers
Vue.use(VueWorker)

Vue.config.productionTip = false

new Vue({
  router,
  store,
  vuetify,
  render: (h) => h(App),
}).$mount("#app")
