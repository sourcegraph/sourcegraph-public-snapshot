import { validateQueryForExhaustiveSearch } from './exhaustive-search-validation'

describe('exhaustive search validation', () => {
    test('works properly with common valid query', () => {
        expect(validateQueryForExhaustiveSearch('context:global insights')).toStrictEqual([])
    })

    test('works properly with valid single rev operator query', () => {
        expect(validateQueryForExhaustiveSearch('context:global rev:* insights')).toStrictEqual([])
    })

    test('[repo:has.content() predicate]', () => {
        expect(validateQueryForExhaustiveSearch('context:global repo:has.content(insights) foo').length).toStrictEqual(
            0
        )
    })

    test('[repo:has.meta() predicate]', () => {
        expect(validateQueryForExhaustiveSearch('context:global repo:has.meta(insights) foo').length).toStrictEqual(0)
    })

    describe('works properly with invalid query', () => {
        test('[multiple rev operator case]', () => {
            expect(validateQueryForExhaustiveSearch('context:global rev:* insights rev:vk').length).toStrictEqual(1)
        })

        test('[has regexp generic pattern]', () => {
            expect(validateQueryForExhaustiveSearch('context:global .* patterntype:regexp').length).toStrictEqual(1)
        })

        test('[file:has.content() predicate]', () => {
            expect(validateQueryForExhaustiveSearch('context:global file:has.content(insights)').length).toStrictEqual(
                1
            )
        })

        test('[file:has.owner() predicate]', () => {
            expect(
                validateQueryForExhaustiveSearch('context:global file:has.owner(insights) foo').length
            ).toStrictEqual(1)
        })

        test('[f:has.contributor() predicate]', () => {
            expect(
                validateQueryForExhaustiveSearch('context:global f:has.contributor(insights) foo').length
            ).toStrictEqual(1)
        })

        test('[f:contains.content() predicate]', () => {
            expect(
                validateQueryForExhaustiveSearch('context:global f:contains.content(insights) foo').length
            ).toStrictEqual(1)
        })

        test('[or operator]', () => {
            expect(validateQueryForExhaustiveSearch('insights or batch-changes').length).toStrictEqual(1)
        })

        test('[and operator]', () => {
            expect(validateQueryForExhaustiveSearch('insights and batch-changes').length).toStrictEqual(1)
        })

        test('[other than type:file]', () => {
            expect(validateQueryForExhaustiveSearch('foo type:file type:diff').length).toStrictEqual(1)
            expect(validateQueryForExhaustiveSearch('foo type:diff').length).toStrictEqual(1)
            expect(validateQueryForExhaustiveSearch('foo type:file type:file')).toStrictEqual([])
            expect(validateQueryForExhaustiveSearch('foo type:file')).toStrictEqual([])
            expect(validateQueryForExhaustiveSearch('foo')).toStrictEqual([])
        })

        test('[all cases combined]', () => {
            expect(
                validateQueryForExhaustiveSearch(
                    'context:global (file:has.content(batch-changes) rev:vk batch) or (patterntype:regexp .* rev:foo)'
                ).length
            ).toStrictEqual(4)
        })
    })
})
