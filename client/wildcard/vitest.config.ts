import { defineProjectWithDefaults } from '../../vitest.shared'

export default defineProjectWithDefaults(__dirname, {
    test: {
        environment: 'jsdom',
        setupFiles: [
            'src/testing/testSetup.test.ts',
            '../shared/dev/reactCleanup.ts',
            '../shared/dev/mockResizeObserver.ts',
            '../shared/dev/mockUniqueId.ts',
        ],
    },
})
