import { expect, test, describe } from 'vitest'

import { getStaticSuffix_TEST_ONLY as getStaticSuffix } from '$root/client/web-sveltekit/src/hooks'

describe('getStaticSuffix', () => {
    test.each([
        ['s1/s2', 's1/s2'],
        ['s1/(group)/s2', 's1/s2'],
        ['s1/[parameter]/s2', 's2'],
        ['s1/[...rest]/s2', 's2'],
        ['s1/[[optional]]/s2', 's2'],
    ])('%s -> %s', (input, expected) => {
        expect(getStaticSuffix(input.split('/')).join('/')).toBe(expected)
    })
})
