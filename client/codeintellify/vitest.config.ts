import { defineProjectWithDefaults } from '../../vitest.shared'

export default defineProjectWithDefaults(__dirname, {
    test: {
        environment: 'jsdom',
        setupFiles: [
            'src/testSetup.test.ts',
            '../testing/src/fetch.js'
        ],
        singleThread: true, // got `failed to terminate worker` occasionally in Bazel CI
    },
})
