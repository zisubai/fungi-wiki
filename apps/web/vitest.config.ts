import { defineConfig } from 'vitest/config';

export default defineConfig({
  test: {
    environment: 'jsdom',
    setupFiles: './src/test/setup.ts',
    clearMocks: true,
    coverage: {
      provider: 'v8',
      include: ['src/api.ts', 'src/App.tsx', 'src/features/**/*.tsx'],
      exclude: ['src/**/*.test.ts', 'src/**/*.test.tsx'],
      reporter: ['text', 'json-summary', 'html'],
      thresholds: {
        statements: 80,
        branches: 55,
        functions: 75,
        lines: 85,
      },
    },
  },
});
