import { jest } from '@jest/globals'
import ResizeObserver from 'resize-observer-polyfill'

if ('ResizeObserver' in window === false) {
    window.ResizeObserver = ResizeObserver
}

jest.mock('use-resize-observer', () => ({
    __esModule: true,

    default: jest.requireActual('use-resize-observer/polyfilled'),
}))
