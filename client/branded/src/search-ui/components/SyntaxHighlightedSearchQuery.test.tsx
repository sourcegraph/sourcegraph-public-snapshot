import { describe, expect, it } from '@jest/globals'
import { render } from '@testing-library/react'

import { SyntaxHighlightedSearchQuery } from './SyntaxHighlightedSearchQuery'

describe('SyntaxHighlightedSearchQuery', () => {
    it('should syntax highlight filter', () => {
        expect(render(<SyntaxHighlightedSearchQuery query="repo:sourcegraph lang:go" />).asFragment()).toMatchSnapshot()
    })

    it('should syntax highlight operator', () => {
        expect(render(<SyntaxHighlightedSearchQuery query="test or spec" />).asFragment()).toMatchSnapshot()
    })

    it('should syntax highlight negated filter', () => {
        expect(render(<SyntaxHighlightedSearchQuery query="-lang:ts test" />).asFragment()).toMatchSnapshot()
    })

    it('should syntax highlight filter and operator', () => {
        expect(
            render(<SyntaxHighlightedSearchQuery query="repo:sourcegraph test and spec" />).asFragment()
        ).toMatchSnapshot()
    })
})
