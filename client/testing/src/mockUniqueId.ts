import { vi } from 'vitest'

// The value returned by the `_.uniqueId()` function depends on the number of previous calls
// because it maintains an internal variable that increments on each call.
// Mock `_.uniqueId()` to avoid updating snapshots once the sequence or number of tests using this helper changes.
vi.mock('lodash', async () => ({
    ...(await vi.importActual<typeof import('lodash')>('lodash')),
    // Keep all actual implementations except uniqueId
    uniqueId: (prefix = '') => `${prefix}test-id`,
}))

// The module factory of `vi.mock()` is not allowed to reference any out-of-scope variables

vi.mock(
    'lodash/uniqueId',
    () =>
        (prefix = '') =>
            `${prefix}test-id`
)
