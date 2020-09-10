import React from 'react'
import { cleanup, render } from '@testing-library/react'
import { SearchPatternType } from '../../graphql-operations'
import { SavedSearchesPanel } from './SavedSearchesPanel'
import { _fetchSavedSearches, authUser } from './utils'

describe('SavedSearchesPanel', () => {
    afterAll(cleanup)

    let container: HTMLElement

    const defaultProps = {
        patternType: SearchPatternType.literal,
        authenticatedUser: authUser,
        fetchSavedSearches: _fetchSavedSearches,
    }

    it('should show all searches by default', () => {
        container = render(<SavedSearchesPanel {...defaultProps} />).container
        const enterpriseHomePanels = container.querySelectorAll('.test-saved-search-entery')
        expect(enterpriseHomePanels.length).toBe(2)
    })
    it('should show one search on "my searches tab"', () => {
        container = render(<SavedSearchesPanel {...defaultProps} mySearchesMode={true} />).container
        const enterpriseHomePanels = container.querySelectorAll('.test-saved-search-entery')
        expect(enterpriseHomePanels.length).toBe(1)
    })
})
