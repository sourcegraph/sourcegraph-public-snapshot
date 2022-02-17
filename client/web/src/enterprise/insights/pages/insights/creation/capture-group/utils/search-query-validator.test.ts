import { searchQueryValidator } from './search-query-validator'

const GOOD_QUERY = 'patterntype:regexp required_version = \\"(.*)\\"  lang:Terraform archived:no fork:no'

const PASSING_VALIDATION = {
    isValidOperator: true,
    isValidPatternType: true,
    isNotRepo: true,
    isNotCommitOrDiff: true,
    isNoNewLines: true,
}

describe('searchQueryValidator', () => {
    it('validates a known good string', () => {
        expect(searchQueryValidator(GOOD_QUERY, true)).toEqual(PASSING_VALIDATION)
    })

    it.each(['and', 'or', 'not'])('validates not containing `%s`', (operator: string) => {
        expect(searchQueryValidator(`${GOOD_QUERY} ${operator}`, true)).toEqual({
            ...PASSING_VALIDATION,
            isValidOperator: false,
        })
    })

    it('validates not using `repo`', () => {
        expect(searchQueryValidator(`${GOOD_QUERY} repo:any`, true)).toEqual({
            ...PASSING_VALIDATION,
            isNotRepo: false,
        })
    })

    it.each(['type:commit', 'type:diff'])('validates not using `commit` or `diff`', (type: string) => {
        expect(searchQueryValidator(`${GOOD_QUERY} ${type}`, true)).toEqual({
            ...PASSING_VALIDATION,
            isNotCommitOrDiff: false,
        })
    })

    it('validates no new lines', () => {
        expect(searchQueryValidator(`${GOOD_QUERY} \\n`, true)).toEqual({
            ...PASSING_VALIDATION,
            isNoNewLines: false,
        })
    })
})
