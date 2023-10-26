import { defineProjectWithDefaults } from '../../vitest.shared'

export default defineProjectWithDefaults(__dirname, {
    test: {
        environmentMatchGlobs: [
            ['**/*.tsx', 'jsdom'],
            ['src/util/(useInputValidation|dom).test.ts', 'jsdom'],
        ],
        setupFiles: ['src/testSetup.test.ts', 'dev/reactCleanup.ts'],
    },
})
