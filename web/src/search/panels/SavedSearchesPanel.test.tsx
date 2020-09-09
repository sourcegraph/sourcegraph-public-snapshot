import React from 'react'
import { cleanup, render } from '@testing-library/react'
import { SearchPatternType } from '../../graphql-operations'
import { SavedSearchesPanel } from './SavedSearchesPanel'
import { _fetchSavedSearches, authUser } from './utils'

// Mock the Monaco input box to make this a shallow test
jest.mock('./SearchPageInput', () => ({
    SearchPageInput: () => null,
}))

describe('SearchPage', () => {
    afterAll(cleanup)

    let container: HTMLElement

    const defaultProps = {
        patternType: SearchPatternType.literal,
        authenticatedUser: authUser,
        fetchSavedSearches: _fetchSavedSearches,
    }

    it('should not show enterprise home panels if on Sourcegraph.com', () => {
        container = render(<SavedSearchesPanel {...defaultProps} />).container
        const enterpriseHomePanels = container.querySelector('.enterprise-home-panels')
        expect(enterpriseHomePanels).toBeFalsy()
    })
})
