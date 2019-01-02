import { validCategories } from './extension'

describe('validCategories', () => {
    test('selects only known categories, sorts, and dedupes', () =>
        expect(validCategories(['Programming languages', 'Other', 'x', 'Other'])).toEqual([
            'Other',
            'Programming languages',
        ]))

    test('returns undefined for undefined', () => expect(validCategories(undefined)).toEqual(undefined))

    test('returns undefined for empty', () => expect(validCategories([])).toEqual(undefined))

    test('returns undefined when no categories are known', () => expect(validCategories(['x'])).toEqual(undefined))
})
