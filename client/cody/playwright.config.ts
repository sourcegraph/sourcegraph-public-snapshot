import { defineConfig } from '@playwright/test'

// eslint-disable-next-line import/no-default-export
export default defineConfig({
    workers: 1,
    testDir: 'test/e2e',
})
