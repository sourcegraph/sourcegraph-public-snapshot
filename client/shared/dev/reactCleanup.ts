import { afterEach } from '@jest/globals'
import { cleanup } from '@testing-library/react'

/**
 * Prevents an issue similar to the one reported here:
 * https://github.com/vitest-dev/vitest/issues/1430
 *
 * Normally `cleanup` should be called automatically after each test.
 * https://testing-library.com/docs/svelte-testing-library/api/#cleanup
 */
afterEach(cleanup)
