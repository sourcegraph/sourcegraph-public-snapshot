import type { PlaywrightTestConfig } from '@playwright/test'

const PORT = process.env.PORT ? Number(process.env.PORT) : 4173

const config: PlaywrightTestConfig = {
    testMatch: '**/*.spec.ts',
    webServer: {
        command: 'pnpm run build:preview && pnpm run preview',
        port: PORT,
        reuseExistingServer: !process.env.CI,
    },
    use: {
        baseURL: `http://localhost:${PORT}`,
    },
}

export default config
