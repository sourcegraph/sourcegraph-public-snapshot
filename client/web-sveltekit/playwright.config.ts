import type { PlaywrightTestConfig } from '@playwright/test'
import { devices, defineConfig } from '@playwright/test'

const PORT = process.env.PORT ? Number(process.env.PORT) : 4173

const config: PlaywrightTestConfig = {
    testMatch: 'src/**/*.spec.ts',
    reporter: 'list',
    // note: if you proxy into a locally running vite preview, you may have to raise this to 60 seconds
    timeout: 60_000,
    expect: {
        timeout: 10_000,
    },
    use: {
        baseURL: `http://localhost:${PORT}`,
    },
    projects: [
        {
            name: 'chromium',
            use: {
                ...devices['Desktop Chrome'],
                launchOptions: {
                    // When in CI, use bazel packaged linux chromium
                    executablePath: process.env.CHROMIUM_BIN,
                },
            },
        },
    ],
}

export default defineConfig(config)
