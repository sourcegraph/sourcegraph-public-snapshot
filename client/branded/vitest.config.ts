import { defineProjectWithDefaults } from '../../vitest.shared'

export default defineProjectWithDefaults(__dirname, {
    test: {
        environment: 'jsdom',
        setupFiles: [
            'src/testSetup.test.ts',
            '../shared/dev/reactCleanup.ts',
            '../shared/dev/mockMatchMedia.ts',
            '../shared/dev/mockUniqueId.ts',
        ],
    },
})
