/* eslint-disable @typescript-eslint/no-var-requires */
/* eslint-disable @typescript-eslint/no-require-imports */

// Use @testing-library/jest-dom in vitest. See
// https://github.com/testing-library/jest-dom/issues/439#issuecomment-1536524120.
import type { TestingLibraryMatchers } from '@testing-library/jest-dom/matchers'
import * as matchers from '@testing-library/jest-dom/matchers'
import { expect } from 'vitest'

declare module 'vitest' {
    interface Assertion<T = any> extends jest.Matchers<void, T>, TestingLibraryMatchers<T, void> {}
}
expect.extend(matchers)

// MessageChannel is not defined in the Vitest jsdom environment.
if (!global.MessageChannel) {
    global.MessageChannel = require('worker_threads').MessageChannel
}
