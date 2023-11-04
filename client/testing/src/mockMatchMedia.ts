import { vi } from 'vitest'

/* eslint-disable ban/ban */

/**
 * Needed for mocking methods not implemented in JSDom, for tests dependent on Monaco editor model.
 */
Object.defineProperty(window, 'matchMedia', {
    writable: true,
    value: vi.fn().mockImplementation(query => ({
        matches: false,
        media: query,
        onchange: null,
        addListener: vi.fn(), // Deprecated
        removeListener: vi.fn(), // Deprecated
        addEventListener: vi.fn(),
        removeEventListener: vi.fn(),
        dispatchEvent: vi.fn(),
    })),
})
