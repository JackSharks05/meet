// Modified by Jack de Haan, 2026 (meet fork of Timeful). See NOTICE.
//
// Third-party analytics (PostHog) has been removed in this fork. To avoid
// editing the many `this.$posthog.*` call sites scattered across the app, this
// plugin installs a no-op `$posthog`: a callable, infinitely-chainable proxy
// that accepts any method call (capture, identify, people.set,
// isFeatureEnabled, …) and does nothing. No analytics data ever leaves the app.

function createNoopProxy() {
  const fn = function () {}
  return new Proxy(fn, {
    get: () => createNoopProxy(),
    apply: () => undefined,
  })
}

export default {
  install(Vue) {
    Vue.prototype.$posthog = createNoopProxy()
  },
}
