import { jest } from '@jest/globals'

// The value returned by the `_.uniqueId()` function depends on the number of previous calls
// because it maintains an internal variable that increments on each call.
// Mock `_.uniqueId()` to avoid updating snapshots once the sequence or number of tests using this helper changes.
jest.mock('lodash', () =>
    Object.assign(jest.requireActual<typeof import('lodash')>('lodash'), {
        // Keep all actual implementations except uniqueId
        uniqueId: (prefix = '') => `${prefix}test-id`,
    })
)

// The module factory of `jest.mock()` is not allowed to reference any out-of-scope variables

jest.mock(
    'lodash/uniqueId',
    () =>
        (prefix = '') =>
            `${prefix}test-id`
)
