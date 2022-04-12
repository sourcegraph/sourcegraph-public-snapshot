import ResizeObserver from 'resize-observer-polyfill'

if ('ResizeObserver' in window === false) {
    window.ResizeObserver = ResizeObserver
}

jest.mock('use-resize-observer', () => ({
    __esModule: true,
    // eslint-disable-next-line @typescript-eslint/no-unsafe-assignment
    default: jest.requireActual('use-resize-observer/polyfilled'),
}))
