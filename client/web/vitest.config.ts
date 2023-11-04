import { defineProjectWithDefaults } from '../../vitest.shared'

export default defineProjectWithDefaults(__dirname, {
    test: {
        environment: 'happy-dom',
        environmentMatchGlobs: [
            ['src/enterprise/code-monitoring/ManageCodeMonitorPage.test.tsx', 'jsdom'], // needs window.confirm
            ['src/enterprise/code-monitoring/CreateCodeMonitorPage.test.tsx', 'jsdom'], // 'Error: Should not already be working.'
            ['src/hooks/useScrollManager/useScrollManager.test.tsx', 'jsdom'], // for correct scroll counting
            ['src/components/KeyboardShortcutsHelp/KeyboardShortcutsHelp.test.tsx', 'jsdom'], // event.getModifierState
        ],

        setupFiles: [
            'src/testSetup.test.ts',
            '../testing/src/reactCleanup.ts',
            '../testing/src/mockMatchMedia.ts',
            '../testing/src/mockUniqueId.ts',
            '../testing/src/mockDate.ts',
        ],
    },
})
