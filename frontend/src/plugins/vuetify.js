// Modified by Jack de Haan, 2026 (meet fork of Timeful). See NOTICE.
// Dark theme inspired by jacksharks05.github.io.
import Vue from "vue"
import Vuetify from "vuetify/lib"
import tailwind from "../../tailwind.config"

Vue.use(Vuetify)

export default new Vuetify({
  theme: {
    dark: true,
    options: { customProperties: true },
    themes: {
      dark: {
        primary: tailwind.theme.colors.green, // availability green (#00994C)
        error: tailwind.theme.colors.red,
        anchor: "#8cbeff",
        background: "#000000",
        surface: "#0a0e1c",
      },
    },
  },
  breakpoint: {
    thresholds: {
      xs: 640,
      sm: 768,
      md: 1024,
      lg: 1280,
    },
    scrollBarWidth: 0,
  },
})
