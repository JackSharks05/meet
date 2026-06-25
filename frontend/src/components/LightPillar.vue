<!--
  Modified by Jack de Haan, 2026 (meet fork of Timeful). See NOTICE.
  Ported to Vue 2 from the React `LightPillar` component on jacksharks05.github.io
  (a Three.js raymarched light-pillar shader). Shaders kept verbatim.
-->
<template>
  <div
    v-if="webGLSupported"
    ref="container"
    :class="['light-pillar-container', className]"
    :style="{ mixBlendMode }"
  />
  <div
    v-else
    :class="['light-pillar-fallback', className]"
    :style="{ mixBlendMode }"
  ></div>
</template>

<script>
import * as THREE from "three"

export default {
  name: "LightPillar",

  props: {
    topColor: { type: String, default: "#621F6D" },
    bottomColor: { type: String, default: "#FE0000" },
    intensity: { type: Number, default: 1.0 },
    rotationSpeed: { type: Number, default: 0.3 },
    interactive: { type: Boolean, default: false },
    className: { type: String, default: "" },
    glowAmount: { type: Number, default: 0.005 },
    pillarWidth: { type: Number, default: 3.0 },
    pillarHeight: { type: Number, default: 0.4 },
    noiseIntensity: { type: Number, default: 0.5 },
    mixBlendMode: { type: String, default: "screen" },
    pillarRotation: { type: Number, default: 0 },
    quality: { type: String, default: "high" },
  },

  data() {
    return { webGLSupported: true }
  },

  mounted() {
    const testCanvas = document.createElement("canvas")
    const gl =
      testCanvas.getContext("webgl") ||
      testCanvas.getContext("experimental-webgl")
    if (!gl) {
      this.webGLSupported = false
      return
    }
    this.$nextTick(() => this.initScene())
  },

  beforeDestroy() {
    this.cleanup()
  },

  methods: {
    initScene() {
      const container = this.$refs.container
      if (!container) return

      const width = container.clientWidth
      const height = container.clientHeight

      const isMobile =
        /Android|webOS|iPhone|iPad|iPod|BlackBerry|IEMobile|Opera Mini/i.test(
          navigator.userAgent,
        )
      const isLowEndDevice =
        isMobile ||
        (navigator.hardwareConcurrency && navigator.hardwareConcurrency <= 4)

      let effectiveQuality = this.quality
      if (isLowEndDevice && this.quality === "high") effectiveQuality = "medium"
      if (isMobile && this.quality !== "low") effectiveQuality = "low"

      const qualitySettings = {
        low: { iterations: 24, waveIterations: 1, pixelRatio: 0.5, precision: "mediump", stepMultiplier: 1.5 },
        medium: { iterations: 40, waveIterations: 2, pixelRatio: 0.65, precision: "mediump", stepMultiplier: 1.2 },
        high: { iterations: 80, waveIterations: 4, pixelRatio: Math.min(window.devicePixelRatio, 2), precision: "highp", stepMultiplier: 1.0 },
      }
      const settings = qualitySettings[effectiveQuality] || qualitySettings.medium
      this.targetFPS = effectiveQuality === "low" ? 30 : 60
      this.frameTime = 1000 / this.targetFPS

      const scene = new THREE.Scene()
      const camera = new THREE.OrthographicCamera(-1, 1, 1, -1, 0, 1)

      let renderer
      try {
        renderer = new THREE.WebGLRenderer({
          antialias: false,
          alpha: true,
          powerPreference: effectiveQuality === "high" ? "high-performance" : "low-power",
          precision: settings.precision,
          stencil: false,
          depth: false,
        })
      } catch (e) {
        this.webGLSupported = false
        return
      }

      renderer.setSize(width, height)
      renderer.setPixelRatio(settings.pixelRatio)
      container.appendChild(renderer.domElement)

      const parseColor = (hex) => {
        const color = new THREE.Color(hex)
        return new THREE.Vector3(color.r, color.g, color.b)
      }

      const vertexShader = `
        varying vec2 vUv;
        void main() {
          vUv = uv;
          gl_Position = vec4(position, 1.0);
        }
      `

      const fragmentShader = `
        precision ${settings.precision} float;

        uniform float uTime;
        uniform vec2 uResolution;
        uniform vec2 uMouse;
        uniform vec3 uTopColor;
        uniform vec3 uBottomColor;
        uniform float uIntensity;
        uniform bool uInteractive;
        uniform float uGlowAmount;
        uniform float uPillarWidth;
        uniform float uPillarHeight;
        uniform float uNoiseIntensity;
        uniform float uRotCos;
        uniform float uRotSin;
        uniform float uPillarRotCos;
        uniform float uPillarRotSin;
        uniform float uWaveSin;
        uniform float uWaveCos;
        varying vec2 vUv;

        const float STEP_MULT = ${settings.stepMultiplier.toFixed(1)};
        const int MAX_ITER = ${settings.iterations};
        const int WAVE_ITER = ${settings.waveIterations};

        void main() {
          vec2 uv = (vUv * 2.0 - 1.0) * vec2(uResolution.x / uResolution.y, 1.0);
          uv = vec2(uPillarRotCos * uv.x - uPillarRotSin * uv.y, uPillarRotSin * uv.x + uPillarRotCos * uv.y);

          vec3 ro = vec3(0.0, 0.0, -10.0);
          vec3 rd = normalize(vec3(uv, 1.0));

          float rotC = uRotCos;
          float rotS = uRotSin;
          if(uInteractive && (uMouse.x != 0.0 || uMouse.y != 0.0)) {
            float a = uMouse.x * 6.283185;
            rotC = cos(a);
            rotS = sin(a);
          }

          vec3 col = vec3(0.0);
          float t = 0.1;

          for(int i = 0; i < MAX_ITER; i++) {
            vec3 p = ro + rd * t;
            p.xz = vec2(rotC * p.x - rotS * p.z, rotS * p.x + rotC * p.z);

            vec3 q = p;
            q.y = p.y * uPillarHeight + uTime;

            float freq = 1.0;
            float amp = 1.0;
            for(int j = 0; j < WAVE_ITER; j++) {
              q.xz = vec2(uWaveCos * q.x - uWaveSin * q.z, uWaveSin * q.x + uWaveCos * q.z);
              q += cos(q.zxy * freq - uTime * float(j) * 2.0) * amp;
              freq *= 2.0;
              amp *= 0.5;
            }

            float d = length(cos(q.xz)) - 0.2;
            float bound = length(p.xz) - uPillarWidth;
            float k = 4.0;
            float h = max(k - abs(d - bound), 0.0);
            d = max(d, bound) + h * h * 0.0625 / k;
            d = abs(d) * 0.15 + 0.01;

            float grad = clamp((15.0 - p.y) / 30.0, 0.0, 1.0);
            col += mix(uBottomColor, uTopColor, grad) / d;

            t += d * STEP_MULT;
            if(t > 50.0) break;
          }

          float widthNorm = uPillarWidth / 3.0;
          col = tanh(col * uGlowAmount / widthNorm);

          col -= fract(sin(dot(gl_FragCoord.xy, vec2(12.9898, 78.233))) * 43758.5453) / 15.0 * uNoiseIntensity;

          gl_FragColor = vec4(col * uIntensity, 1.0);
        }
      `

      const pillarRotRad = (this.pillarRotation * Math.PI) / 180
      const waveSin = Math.sin(0.4)
      const waveCos = Math.cos(0.4)

      this.mouse = new THREE.Vector2(0, 0)
      const material = new THREE.ShaderMaterial({
        vertexShader,
        fragmentShader,
        uniforms: {
          uTime: { value: 0 },
          uResolution: { value: new THREE.Vector2(width, height) },
          uMouse: { value: this.mouse },
          uTopColor: { value: parseColor(this.topColor) },
          uBottomColor: { value: parseColor(this.bottomColor) },
          uIntensity: { value: this.intensity },
          uInteractive: { value: this.interactive },
          uGlowAmount: { value: this.glowAmount },
          uPillarWidth: { value: this.pillarWidth },
          uPillarHeight: { value: this.pillarHeight },
          uNoiseIntensity: { value: this.noiseIntensity },
          uRotCos: { value: 1.0 },
          uRotSin: { value: 0.0 },
          uPillarRotCos: { value: Math.cos(pillarRotRad) },
          uPillarRotSin: { value: Math.sin(pillarRotRad) },
          uWaveSin: { value: waveSin },
          uWaveCos: { value: waveCos },
        },
        transparent: true,
        depthWrite: false,
        depthTest: false,
      })

      const geometry = new THREE.PlaneGeometry(2, 2)
      const mesh = new THREE.Mesh(geometry, material)
      scene.add(mesh)

      this.scene = scene
      this.camera = camera
      this.renderer = renderer
      this.material = material
      this.geometry = geometry
      this.time = 0

      // Interaction
      this.handleMouseMove = (event) => {
        if (!this.interactive) return
        const rect = container.getBoundingClientRect()
        const x = ((event.clientX - rect.left) / rect.width) * 2 - 1
        const y = -((event.clientY - rect.top) / rect.height) * 2 + 1
        this.mouse.set(x, y)
      }
      if (this.interactive) {
        container.addEventListener("mousemove", this.handleMouseMove, { passive: true })
      }

      // Animation
      this.lastTime = performance.now()
      const animate = (currentTime) => {
        if (!this.material || !this.renderer || !this.scene || !this.camera) return
        const deltaTime = currentTime - this.lastTime
        if (deltaTime >= this.frameTime) {
          this.time += 0.016 * this.rotationSpeed
          const tm = this.time
          this.material.uniforms.uTime.value = tm
          this.material.uniforms.uRotCos.value = Math.cos(tm * 0.3)
          this.material.uniforms.uRotSin.value = Math.sin(tm * 0.3)
          this.renderer.render(this.scene, this.camera)
          this.lastTime = currentTime - (deltaTime % this.frameTime)
        }
        this.raf = requestAnimationFrame(animate)
      }
      this._animate = animate

      this.handleVisibilityChange = () => {
        if (document.hidden) this.stopRaf()
        else this.startRaf()
      }
      document.addEventListener("visibilitychange", this.handleVisibilityChange)
      this.startRaf()

      // Resize
      this.handleResize = () => {
        if (this._resizeTimeout) window.clearTimeout(this._resizeTimeout)
        this._resizeTimeout = window.setTimeout(() => {
          if (!this.renderer || !this.material || !this.$refs.container) return
          const w = this.$refs.container.clientWidth
          const h = this.$refs.container.clientHeight
          this.renderer.setSize(w, h)
          this.material.uniforms.uResolution.value.set(w, h)
        }, 150)
      }
      window.addEventListener("resize", this.handleResize, { passive: true })
    },

    startRaf() {
      if (!this.raf && this._animate) {
        this.lastTime = performance.now()
        this.raf = requestAnimationFrame(this._animate)
      }
    },

    stopRaf() {
      if (this.raf) {
        cancelAnimationFrame(this.raf)
        this.raf = null
      }
    },

    cleanup() {
      if (this.handleVisibilityChange)
        document.removeEventListener("visibilitychange", this.handleVisibilityChange)
      if (this.handleResize) window.removeEventListener("resize", this.handleResize)
      if (this.interactive && this.handleMouseMove && this.$refs.container)
        this.$refs.container.removeEventListener("mousemove", this.handleMouseMove)
      this.stopRaf()

      if (this.renderer) {
        this.renderer.dispose()
        this.renderer.forceContextLoss()
        const el = this.renderer.domElement
        if (this.$refs.container && el && this.$refs.container.contains(el)) {
          this.$refs.container.removeChild(el)
        }
      }
      if (this.material) this.material.dispose()
      if (this.geometry) this.geometry.dispose()

      this.renderer = null
      this.material = null
      this.scene = null
      this.camera = null
      this.geometry = null
      this.raf = null
      this._animate = null
    },
  },
}
</script>

<!-- Non-scoped: the renderer's <canvas> is appended at runtime and would not
     receive scoped attributes. Class names are specific enough. -->
<style>
.light-pillar-container {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  pointer-events: none;
}
.light-pillar-container canvas {
  width: 100% !important;
  height: 100% !important;
  display: block;
}
.light-pillar-fallback {
  position: absolute;
  inset: 0;
}
</style>
