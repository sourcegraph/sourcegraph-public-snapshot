import React from 'react'
import { mount } from 'enzyme'
import { SyntaxHighlightedSearchQuery } from './SyntaxHighlightedSearchQuery'

describe('SyntaxHighlightedSearchQuery', () => {
    it('should syntax highlight filter', () => {
        expect(mount(<SyntaxHighlightedSearchQuery query="repo:sourcegraph lang:go" />)).toMatchSnapshot()
    })

    it('should syntax highlight operator', () => {
        expect(mount(<SyntaxHighlightedSearchQuery query="test or spec" />)).toMatchSnapshot()
    })

    it('should syntax highlight negated filter', () => {
        expect(mount(<SyntaxHighlightedSearchQuery query="-lang:ts test" />)).toMatchSnapshot()
    })

    it('should syntax highlight filter and operator', () => {
        expect(mount(<SyntaxHighlightedSearchQuery query="repo:sourcegraph test and spec" />)).toMatchSnapshot()
    })
})
