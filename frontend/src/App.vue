<!-- Modified by Jack de Haan, 2026 (meet fork of Timeful). See NOTICE. -->
<template>
  <v-app>
    <div class="jdh-backdrop">
      <LightPillar
        :intensity="1.0"
        :rotation-speed="0.25"
        :pillar-rotation="90"
      />
    </div>
    <AutoSnackbar color="error" :text="error" />
    <AutoSnackbar color="tw-bg-blue" :text="info" />
    <SignInNotSupportedDialog v-model="webviewDialog" />
    <SignInDialog v-model="signInDialog" @signIn="_signIn" />
    <NewDialog
      v-if="isAdminBuild"
      v-model="newDialogOptions.show"
      :type="newDialogOptions.openNewGroup ? 'group' : 'event'"
      :contactsPayload="newDialogOptions.contactsPayload"
      :no-tabs="newDialogOptions.eventOnly"
      :folder-id="newDialogOptions.folderId"
    />
    <UpgradeDialog
      :value="upgradeDialogVisible"
      @input="handleUpgradeDialogInput"
    />
    <div
      v-if="showHeader"
      class="jdh-header tw-fixed tw-z-40 tw-h-14 tw-w-screen sm:tw-h-16"
    >
      <div
        class="tw-relative tw-m-auto tw-flex tw-h-full tw-max-w-6xl tw-items-center tw-px-4"
      >
        <router-link :to="{ name: 'home' }" class="jdh-wordmark">
          meet with jdh
        </router-link>

        <v-spacer />

        <!-- Event creation + account UI is operator-only (admin build). Public
             poll links are respond-only and show just the wordmark. -->
        <template v-if="isAdminBuild">
          <!-- Show "Create an event" on every admin route except home (which has
               its own "+ Create new" button) — including the main/landing page. -->
          <v-btn
            v-if="$route.name !== 'home'"
            id="top-right-create-btn"
            text
            @click="() => _createNew(true)"
          >
            Create an event
          </v-btn>
          <v-btn
            v-if="$route.name === 'home' && !isPhone"
            color="primary"
            class="tw-mx-2 tw-rounded-md"
            :style="{
              boxShadow: '0px 2px 8px 0px #00994C80 !important',
            }"
            @click="() => _createNew()"
          >
            + Create new
          </v-btn>
          <!-- No meet sign-in: the admin build is operator-only by tailnet
               isolation, so there's no sign-in button. AuthUserMenu still shows
               if a session somehow exists, but it's never required. -->
          <div v-if="authUser" class="sm:tw-ml-4">
            <AuthUserMenu />
          </div>
        </template>
      </div>
    </div>

    <v-main>
      <div class="tw-flex tw-h-screen tw-flex-col">
        <div
          class="tw-relative tw-flex-1 tw-overscroll-auto"
          :class="routerViewClass"
        >
          <router-view v-if="loaded" :key="$route.fullPath" />
        </div>
      </div>
    </v-main>
  </v-app>
</template>

<style>
html {
  overflow-y: auto !important;
  scroll-behavior: smooth;
}

* {
  font-family: "Spectral", ui-serif, Georgia, "Times New Roman", Times, serif;
}

/* Full-viewport LightPillar backdrop behind all content */
.jdh-backdrop {
  position: fixed;
  inset: 0;
  z-index: 0;
  pointer-events: none;
  background: #000;
}

/* Content layers above the backdrop */
.v-application .v-main {
  position: relative;
  z-index: 1;
}

/* Glassy dark header (jacksharks05 surface) */
.jdh-header {
  background: rgba(10, 14, 28, 0.6);
  backdrop-filter: blur(10px);
  border-bottom: 1px solid var(--jdh-outline);
  z-index: 40;
}
.jdh-wordmark {
  font-size: 22px;
  font-weight: 600;
  letter-spacing: 0.5px;
  color: #fff !important;
  text-decoration: none;
}

.v-messages__message {
  font-size: theme("fontSize.xs");
  line-height: 1.25;
}
.v-input--selection-controls {
  margin-top: 0px !important;
  padding-top: 0px !important;
}

/** Buttons */
.v-btn {
  letter-spacing: unset !important;
  text-transform: unset !important;
}
.v-btn:not(.v-btn--round, .v-btn-toggle > .v-btn).v-size--default {
  height: 38px !important;
  border-radius: theme("borderRadius.md") !important;
}

.v-btn.v-btn--is-elevated {
  -webkit-box-shadow: 0px 2px 6px 0px rgba(0, 0, 0, 0.15) !important;
  -moz-box-shadow: 0px 2px 6px 0px rgba(0, 0, 0, 0.15) !important;
  box-shadow: 0px 2px 6px 0px rgba(0, 0, 0, 0.15) !important;
  border: 1px solid theme("colors.light-gray-stroke");
}

.v-btn.v-btn--is-elevated.tw-bg-white {
  -webkit-box-shadow: 0px 1px 4px 0px rgba(0, 0, 0, 0.25) !important;
  -moz-box-shadow: 0px 1px 4px 0px rgba(0, 0, 0, 0.25) !important;
  box-shadow: 0px 1px 4px 0px rgba(0, 0, 0, 0.25) !important;
  border: 1px solid theme("colors.off-white");
}

.v-btn.v-btn--is-elevated.primary,
.v-btn.v-btn--is-elevated.tw-bg-green,
.v-btn.v-btn--is-elevated.tw-bg-white.tw-text-green {
  -webkit-box-shadow: 0px 2px 8px 0px #00994c80 !important;
  -moz-box-shadow: 0px 2px 8px 0px #00994c80 !important;
  box-shadow: 0px 2px 8px 0px #00994c80 !important;
  border: 1px solid theme("colors.light-green") !important;
}

.v-btn.v-btn--is-elevated.tw-bg-very-dark-gray {
  -webkit-box-shadow: 0px 2px 6px 0px rgba(0, 0, 0, 0.25) !important;
  -moz-box-shadow: 0px 2px 6px 0px rgba(0, 0, 0, 0.25) !important;
  box-shadow: 0px 2px 6px 0px rgba(0, 0, 0, 0.25) !important;
  border: 1px solid theme("colors.dark-gray") !important;
}

.v-btn.v-btn--is-elevated.tw-bg-blue,
.v-btn.v-btn--is-elevated.tw-bg-white.tw-text-blue {
  -webkit-box-shadow: 0px 2px 6px 0px rgba(0, 0, 0, 0.25) !important;
  -moz-box-shadow: 0px 2px 6px 0px rgba(0, 0, 0, 0.25) !important;
  box-shadow: 0px 2px 6px 0px rgba(0, 0, 0, 0.25) !important;
  border: 1px solid theme("colors.light-blue") !important;
}

/** Drop shadows */
.v-text-field.v-text-field--solo:not(.v-text-field--solo-flat)
  > .v-input__control
  > .v-input__slot {
  filter: drop-shadow(0 0.5px 2px rgba(0, 0, 0, 0.1)) !important;
  box-shadow: inset 0 -1px 0 0 rgba(0, 0, 0, 0.1) !important;
  border-radius: theme("borderRadius.md") !important;
  border: 1px solid #4f4f4f1f !important;
}
.v-menu__content {
  box-shadow: 0px 5px 5px -1px rgba(0, 0, 0, 0.1),
    0px 8px 10px 0.5px rgba(0, 0, 0, 0.07), 0px 3px 14px 1px rgba(0, 0, 0, 0.06) !important;
}
.overlay-avail-shadow-green {
  box-shadow: 0px 3px 6px 0px #1c7d454d !important;
}
.overlay-avail-shadow-yellow {
  box-shadow: 0px 2px 8px 0px #e5a8004d !important;
}

/** Switch  */
.v-input--switch--inset .v-input--selection-controls__input {
  margin-right: 0 !important;
  transform: scale(80%) !important;
}
.v-input--switch__track.primary--text {
  border: 2px theme("colors.light-green") solid !important;
}
.v-input--switch__track {
  border: 2px theme("colors.gray") solid !important;
  background-color: theme("colors.gray") !important;
  box-shadow: 0px 0.74px 4.46px 0px rgba(0, 0, 0, 0.1) !important;
}
.v-input--is-label-active .v-input--switch__track {
  background-color: currentColor !important;
  box-shadow: 0px 1.5px 4.5px 0px rgba(0, 0, 0, 0.2) !important;
}
.v-input--switch--inset .v-input--switch__track,
.v-input--switch--inset .v-input--selection-controls__input {
  opacity: 1 !important;
}
.v-input--switch__thumb {
  background-color: white !important;
}
.v-text-field__details {
  padding: 0 !important;
}

/** Error color */
.error--text .v-input__slot {
  outline: red solid;
  border-radius: 3px;
}
</style>

<script>
import { mapMutations, mapState, mapActions, mapGetters } from "vuex"
import {
  get,
  isPhone,
  signInGoogle,
  signInOutlook,
  isPremiumUser,
} from "@/utils"
import {
  authTypes,
  calendarTypes,
  eventTypes,
  numFreeEvents,
  upgradeDialogTypes,
} from "@/constants"
import AutoSnackbar from "@/components/AutoSnackbar"
import AuthUserMenu from "@/components/AuthUserMenu.vue"
import SignInNotSupportedDialog from "@/components/SignInNotSupportedDialog.vue"
import isWebview from "is-ua-webview"
import NewDialog from "./components/NewDialog.vue"
import UpgradeDialog from "@/components/pricing/UpgradeDialog.vue"
import SignInDialog from "@/components/SignInDialog.vue"
import LightPillar from "@/components/LightPillar.vue"

export default {
  name: "App",

  metaInfo: {
    htmlAttrs: {
      lang: "en-US",
    },
  },

  components: {
    AutoSnackbar,
    AuthUserMenu,
    SignInNotSupportedDialog,
    NewDialog,
    UpgradeDialog,
    SignInDialog,
    LightPillar,
  },

  data: () => ({
    mounted: false,
    loaded: false,
    scrollY: 0,
    webviewDialog: false,
    signInDialog: false,
  }),

  computed: {
    ...mapGetters(["isPremiumUser"]),
    ...mapState([
      "authUser",
      "error",
      "info",
      "enablePaywall",
      "upgradeDialogVisible",
      "newDialogOptions",
    ]),
    isPhone() {
      return isPhone(this.$vuetify)
    },
    isAdminBuild() {
      // Event creation + account UI is operator-only. The public (respond-only)
      // build leaves VUE_APP_ADMIN unset; the admin build sets it to "true".
      return process.env.VUE_APP_ADMIN === "true"
    },
    showHeader() {
      return (
        this.$route.name !== "landing" &&
        this.$route.name !== "auth" &&
        this.$route.name !== "privacy-policy"
      )
    },
    showFeedbackBtn() {
      return !this.isPhone || this.$route.name === "home"
    },
    routerViewClass() {
      let c = ""
      if (this.showHeader) {
        if (this.isPhone) {
          c += "tw-pt-12 "
        } else {
          c += "tw-pt-14 "
        }
      }
      return c
    },
  },

  methods: {
    ...mapMutations([
      "setAuthUser",
      "setSignUpFormEnabled",
      "setPricingPageConversion",
      "setEnablePaywall",
      "setFeatureFlagsLoaded",
    ]),
    ...mapActions([
      "getEvents",
      "showUpgradeDialog",
      "hideUpgradeDialog",
      "createNew",
    ]),
    handleScroll(e) {
      this.scrollY = window.scrollY
    },
    _createNew(eventOnly = false) {
      this.$posthog.capture("create_new_button_clicked", {
        eventOnly: eventOnly,
      })
      this.createNew({ eventOnly })
    },
    signIn() {
      if (
        this.$route.name === "event" ||
        this.$route.name === "group" ||
        this.$route.name === "signUp"
      ) {
        if (isWebview(navigator.userAgent)) {
          this.webviewDialog = true
          return
        }
        this.signInDialog = true
      }
    },
    _signIn(calendarType) {
      if (
        this.$route.name === "event" ||
        this.$route.name === "group" ||
        this.$route.name === "signUp"
      ) {
        let state
        if (this.$route.name === "event") {
          state = {
            eventId: this.$route.params.eventId,
            type: authTypes.EVENT_SIGN_IN,
          }
        } else if (this.$route.name === "group") {
          state = {
            groupId: this.$route.params.groupId,
            type: authTypes.GROUP_SIGN_IN,
          }
        }
        if (calendarType === calendarTypes.GOOGLE) {
          signInGoogle({
            state,
            selectAccount: true,
          })
        } else if (calendarType === calendarTypes.OUTLOOK) {
          signInOutlook({
            state,
            selectAccount: true,
          })
        }
      }
    },
    setFeatureFlags() {
      if (!this.$posthog) return

      // this.setSignUpFormEnabled(this.$posthog.isFeatureEnabled("sign-up-form"))
      // this.setPricingPageConversion(
      // this.$posthog.getFeatureFlag("pricing-page-conversion")
      // )
      // )
      // this.setEnablePaywall(this.$posthog.isFeatureEnabled("enable-paywall"))
      this.setFeatureFlagsLoaded(true)
    },
    trackFeedbackClick() {
      this.$posthog.capture("give_feedback_button_clicked")
    },
    handleUpgradeDialogInput(value) {
      if (!value) {
        this.hideUpgradeDialog()
      }
    },
  },

  async created() {
    await get("/user/profile")
      .then((authUser) => {
        this.setAuthUser(authUser)

        this.$posthog?.identify(authUser._id, {
          email: authUser.email,
          firstName: authUser.firstName,
          lastName: authUser.lastName,
        })
      })
      .catch(() => {
        this.setAuthUser(null)
      })
      .finally(() => {
        this.loaded = true
      })

    // Event listeners
    window.addEventListener("scroll", this.handleScroll)

    this.getEvents()
  },

  mounted() {
    this.mounted = true
    this.scrollY = window.scrollY
  },

  beforeDestroy() {
    window.removeEventListener("scroll", this.handleScroll)
  },

  watch: {
    $route: {
      immediate: true,
      handler() {
        if (this.$route.name) {
          this.$posthog?.capture("$pageview")
        }
      },
    },
    authUser: {
      immediate: true,
      handler() {
        if (this.$posthog) {
          this.setFeatureFlags()
          // Check feature flags (only if posthog is enabled)
          // this.$posthog.setPersonPropertiesForFlags({
          //   email: this.authUser?.email,
          // })
          // this.$posthog.onFeatureFlags(() => {
          //   this.setFeatureFlags()
          // })
        }
      },
    },
  },
}
</script>
