import { vi } from 'vitest'

// Inspired by React's packages/dom-event-testing-library/domEnvironment.js

const originalPlatform = window.navigator.platform
const platformGetter = vi.spyOn(window.navigator, 'platform', 'get')

/**
 * Change environment host platform.
 */
export const platform = {
    reset() {
        platformGetter.mockReturnValue(originalPlatform)
    },
    set(name: 'mac' | 'windows') {
        switch (name) {
            case 'mac': {
                platformGetter.mockReturnValue('MacIntel')
                break
            }
            case 'windows': {
                platformGetter.mockReturnValue('Win32')
                break
            }
            default: {
                break
            }
        }
    },
}
