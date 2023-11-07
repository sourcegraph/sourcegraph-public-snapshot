import { describe, expect, it } from '@jest/globals'
import { screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import sinon from 'sinon'

import type { SearchScope } from '@sourcegraph/shared/src/schema/settings.schema'
import type { Filter } from '@sourcegraph/shared/src/search/stream'
import { renderWithBrandedContext } from '@sourcegraph/wildcard/src/testing'

import { getDynamicFilterLinks, getRepoFilterLinks, getSearchSnippetLinks } from './FilterLink'

describe('FilterLink', () => {
    const repoFilter1: Filter = {
        label: 'gitlab.com/sourcegraph/sourcegraph',
        value: 'repo:^gitlab\\.com/sourcegraph/sourcgreaph$',
        count: 5,
        limitHit: false,
        kind: 'repo',
    }

    const repoFilter2: Filter = {
        label: 'github.com/microsoft/vscode',
        value: 'repo:^github\\.com/microsoft/vscode$',
        count: 201,
        limitHit: true,
        kind: 'repo',
    }

    const langFilter1: Filter = {
        label: 'lang:go',
        value: 'lang:go',
        count: 500,
        limitHit: true,
        kind: 'lang',
    }

    const langFilter2: Filter = {
        label: 'lang:typescript',
        value: 'lang:typescript',
        count: 241,
        limitHit: false,
        kind: 'lang',
    }

    const fileFilter: Filter = {
        label: '-file:_test\\.go$',
        value: '-file:_test\\.go$',
        count: 1,
        limitHit: false,
        kind: 'file',
    }

    it('should have correct links for repos', () => {
        const filters: Filter[] = [repoFilter1, langFilter1, repoFilter2, langFilter2, fileFilter]
        const onFilterChosen = sinon.stub()

        const links = getRepoFilterLinks(filters, onFilterChosen)
        expect(links).toHaveLength(2)
        expect(renderWithBrandedContext(<>{links}</>).asFragment()).toMatchSnapshot()
    })

    it('should have show icons for repos on cloud', () => {
        const filters: Filter[] = [repoFilter1, langFilter1, repoFilter2, langFilter2, fileFilter]
        const onFilterChosen = sinon.stub()

        const links = getRepoFilterLinks(filters, onFilterChosen)
        expect(links).toHaveLength(2)

        const { asFragment } = renderWithBrandedContext(<>{links}</>)
        expect(screen.getByLabelText('github.com')).toBeInTheDocument()
        expect(screen.getByLabelText('gitlab.com')).toBeInTheDocument()
        expect(asFragment()).toMatchSnapshot()
    })

    it('should have no repo links if no repo filters present', () => {
        const filters: Filter[] = [langFilter1, langFilter2, fileFilter]
        const onFilterChosen = sinon.stub()

        const links = getRepoFilterLinks(filters, onFilterChosen)
        expect(links).toHaveLength(0)
    })

    it('should have correct links for dynamic filters', () => {
        const filters: Filter[] = [repoFilter1, langFilter1, repoFilter2, langFilter2, fileFilter]
        const onFilterChosen = sinon.stub()

        const links = getDynamicFilterLinks(filters, ['file', 'lang', 'utility'], onFilterChosen)
        expect(links).toHaveLength(3)
        expect(renderWithBrandedContext(<>{links}</>).asFragment()).toMatchSnapshot()
    })

    it('should have no dynamic filters links if no dynamic filters present', () => {
        const filters: Filter[] = [repoFilter1, repoFilter2]
        const onFilterChosen = sinon.stub()

        const links = getDynamicFilterLinks(filters, ['file', 'lang', 'utility'], onFilterChosen)
        expect(links).toHaveLength(0)
    })

    it('should have correct links for scopes', () => {
        const scopes: SearchScope[] = [
            {
                name: 'This is a search scope with a very long name lorem ipsum dolor sit amet',
                value: 'repo:sourcegraph',
            },
            { name: 'All results', value: 'count:all' },
        ]
        const onFilterChosen = sinon.stub()

        const links = getSearchSnippetLinks({ subjects: [], final: { 'search.scopes': scopes } }, onFilterChosen)
        expect(links).toHaveLength(2)
        expect(renderWithBrandedContext(<>{links}</>).asFragment()).toMatchSnapshot()
    })

    it('should have no snippet links if no snippets present', () => {
        const onFilterChosen = sinon.stub()

        const links = getSearchSnippetLinks({ subjects: [], final: {} }, onFilterChosen)
        expect(links).toHaveLength(0)
    })

    it('should call correct callback when clicked', () => {
        const filters: Filter[] = [repoFilter1]
        const onFilterChosen = sinon.spy()

        const links = getRepoFilterLinks(filters, onFilterChosen)
        renderWithBrandedContext(<>{links}</>)
        userEvent.click(screen.getByTestId('filter-link'))

        sinon.assert.calledWithExactly(onFilterChosen, repoFilter1.value, repoFilter1.kind)
    })
})
