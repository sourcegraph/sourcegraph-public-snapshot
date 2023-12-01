// TODO(sqs): for some reason, `import '@sourcegraph/testing/src/jestDomMatchers'` does not work (it
// does not extend Assertion with the types).

import type { TestingLibraryMatchers } from '@testing-library/jest-dom/matchers'
import * as matchers from '@testing-library/jest-dom/matchers'
import { expect } from 'vitest'

declare module 'vitest' {
    interface Assertion<T = any> extends jest.Matchers<void, T>, TestingLibraryMatchers<T, void> {}
}
expect.extend(matchers)
