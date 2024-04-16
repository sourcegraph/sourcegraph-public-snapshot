import { defineProjectWithDefaults } from '../../vitest.shared'

export default defineProjectWithDefaults(__dirname, {
    test: {
        environment: 'jsdom',
        setupFiles: [
            'src/testing/testSetup.test.ts',
            '../testing/src/reactCleanup.ts',
            '../testing/src/mockResizeObserver.ts',
            '../testing/src/mockUniqueId.ts',
        ],
    },
})
