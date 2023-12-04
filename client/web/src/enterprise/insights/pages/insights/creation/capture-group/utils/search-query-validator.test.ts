import { describe, expect, it } from 'vitest'

import { searchQueryValidator } from './search-query-validator'

const GOOD_QUERY = 'patterntype:regexp required_version = \\"(.*)\\"  lang:Terraform archived:no fork:no'

const PASSING_VALIDATION = {
    isValidOperator: true,
    isValidPatternType: true,
    isNotRepo: true,
    isNotContext: true,
    isNotCommitOrDiff: true,
    isNoNewLines: true,
    isNotRev: true,
}

describe('searchQueryValidator', () => {
    it('validates a known good string', () => {
        expect(searchQueryValidator(GOOD_QUERY)).toEqual(PASSING_VALIDATION)
    })

    it.each(['and', 'or', 'not'])('validates not containing `%s`', (operator: string) => {
        expect(searchQueryValidator(`${GOOD_QUERY} ${operator}`)).toEqual({
            ...PASSING_VALIDATION,
            isValidOperator: false,
        })
    })

    it('validates not using `repo`', () => {
        expect(searchQueryValidator(`${GOOD_QUERY} repo:any`)).toEqual({
            ...PASSING_VALIDATION,
            isNotRepo: false,
        })
    })

    it.each(['type:commit', 'type:diff'])('validates not using `commit` or `diff`', (type: string) => {
        expect(searchQueryValidator(`${GOOD_QUERY} ${type}`)).toEqual({
            ...PASSING_VALIDATION,
            isNotCommitOrDiff: false,
        })
    })

    it('validates no new lines', () => {
        expect(searchQueryValidator(`${GOOD_QUERY} \\n`)).toEqual({
            ...PASSING_VALIDATION,
            isNoNewLines: false,
        })
    })

    it('validates not using `rev`', () => {
        expect(searchQueryValidator(`${GOOD_QUERY} rev:any`)).toEqual({
            ...PASSING_VALIDATION,
            isNotRev: false,
        })
    })
})
