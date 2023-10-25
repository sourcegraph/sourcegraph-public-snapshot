import ResizeObserver from 'resize-observer-polyfill'
import { vi } from 'vitest'

if ('ResizeObserver' in window === false) {
    window.ResizeObserver = ResizeObserver
}

vi.mock('use-resize-observer', () => ({
    __esModule: true,

    default: vi.requireActual('use-resize-observer/polyfilled'),
}))
