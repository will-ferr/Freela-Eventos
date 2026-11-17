import { defineConfig, mergeConfig } from 'vitest/config'
import viteConfig from './vite.config.ts'

export default defineConfig((env) =>
  mergeConfig(viteConfig(env), {
    test: {
      environment: 'jsdom',
      include: ['src/**/*.spec.ts'],
      restoreMocks: true,
      coverage: {
        provider: 'v8',
        include: ['src/**/*.{ts,vue}'],
        exclude: ['src/**/*.spec.ts', 'src/main.ts', 'src/types/**'],
      },
    },
  }),
)
