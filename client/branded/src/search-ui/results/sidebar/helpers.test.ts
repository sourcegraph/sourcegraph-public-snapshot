import { renderHook } from '@testing-library/react'
import { describe, expect, it } from 'vitest'

import { FilterType } from '@sourcegraph/shared/src/search/query/filters'
import type { Filter } from '@sourcegraph/shared/src/search/stream'

import { useLastRepoName } from './helpers'

interface Props {
    query: string
    filters: Filter[]
}

function createRepoFilter(value: string): Filter {
    return {
        kind: FilterType.repo,
        value: `${value}@HEAD`,
        label: value,
        count: 1,
        limitHit: false,
    }
}

function setup(props: Props) {
    return renderHook(({ query, filters }: Props) => useLastRepoName(query, filters), { initialProps: props })
}

describe('useLastRepoName', () => {
    const repoName = 'sourcegraph'
    const repoQuery = 'repo:sourcegraph'
    const initialProps: Props = { query: repoQuery, filters: [createRepoFilter(repoName)] }

    it('returns no repo name if we have never seen any results', () => {
        // Empty query and filters
        const { result, rerender } = setup({ query: '', filters: [] })
        expect(result.current).toBe('')

        // Single repo query
        rerender({ query: 'repo:sourcegraph', filters: [] })
        expect(result.current).toBe('')

        // Multiple repo query
        rerender({ query: 'repo:sourcegraph/a repo:sourcegraph/b', filters: [] })
        expect(result.current).toBe('')
    })

    it('returns the repo name if results contain a single repo', () => {
        const filters = [createRepoFilter(repoName)]

        const { result, rerender } = setup({ query: '', filters })
        expect(result.current).toBe(repoName)

        rerender({ query: 'no repo filter', filters: filters.slice() })
        expect(result.current).toBe(repoName)

        rerender({ query: 'repo:whatever', filters: filters.slice() })
        expect(result.current).toBe(repoName)
    })

    it('returns the previous repo name if the query matches but there are no results', () => {
        const { result, rerender } = setup(initialProps)
        expect(result.current).toBe(repoName)

        rerender({ query: repoQuery, filters: [] })
        expect(result.current).toBe(repoName)
    })

    it('returns an empty repo name if the repo query has changed', () => {
        let { result, rerender } = setup(initialProps)
        expect(result.current).toBe(repoName)

        rerender({ query: 'repo:another', filters: [] })
        expect(result.current).toBe('')

        // reset
        ;({ result, rerender } = setup(initialProps))
        expect(result.current).toBe(repoName)

        rerender({ query: repoQuery + ' repo:anotherone', filters: [] })
        expect(result.current).toBe('')
    })

    it('returns an empty repo name if the results contain multiple repos', () => {
        const { result, rerender } = setup(initialProps)
        expect(result.current).toBe(repoName)

        rerender({ query: repoQuery, filters: [createRepoFilter('sourcegraph/a'), createRepoFilter('sourcegraph/b')] })
        expect(result.current).toBe('')
    })
})
