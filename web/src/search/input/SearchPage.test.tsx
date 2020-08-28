import React from 'react'
import { cleanup, render } from '@testing-library/react'
import { SearchPage, SearchPageProps } from './SearchPage'

jest.mock('../../tracking/eventLogger', () => ({
    eventLogger: { logViewEvent: () => undefined },
}))

jest.mock('./SearchPageInput', () => ({
    SearchPageInput: () => null,
}))

describe('SearchPage', () => {
    afterAll(cleanup)

    let container: HTMLElement

    const defaultProps = {
        isSourcegraphDotCom: false,
        settingsCascade: {
            final: null,
            subjects: null,
        },
        extensionsController: {
            services: {},
        },
        location: {
            pathname: '',
        },
    } as SearchPageProps

    it('should have `with-content-below` class if on Sourcegraph.com', () => {
        container = render(<SearchPage {...defaultProps} isSourcegraphDotCom={true} />).container
        const searchContainer = container.querySelector('.search-page__search-container')
        expect(searchContainer?.classList.contains('search-page__search-container--with-content-below')).toBeTruthy()
    })

    it('should have `with-content-below` class if showEnterpriseHomePanels enabled', () => {
        container = render(<SearchPage {...defaultProps} showEnterpriseHomePanels={true} />).container
        const searchContainer = container.querySelector('.search-page__search-container')
        expect(searchContainer?.classList.contains('search-page__search-container--with-content-below')).toBeTruthy()
    })

    it('should not have `with-content-below` class if showEnterpriseHomePanels disabled and not on Sourcegraph.com', () => {
        container = render(<SearchPage {...defaultProps} />).container
        const searchContainer = container.querySelector('.search-page__search-container')
        expect(searchContainer?.classList.contains('search-page__search-container--with-content-below')).toBeFalsy()
    })

    it('should not show enterprise home panels if on Sourcegraph.com', () => {
        container = render(<SearchPage {...defaultProps} isSourcegraphDotCom={true} />).container
        const enterpriseHomePanels = container.querySelector('.enterprise-home-panels')
        expect(enterpriseHomePanels).toBeFalsy()
    })

    it('should not show enterprise home panels if showEnterpriseHomePanels disabled', () => {
        container = render(<SearchPage {...defaultProps} />).container
        const enterpriseHomePanels = container.querySelector('.enterprise-home-panels')
        expect(enterpriseHomePanels).toBeFalsy()
    })

    it('should show enterprise home panels if showEnterpriseHomePanels enabled and not on Sourcegraph.com', () => {
        container = render(<SearchPage {...defaultProps} showEnterpriseHomePanels={true} />).container
        const enterpriseHomePanels = container.querySelector('.enterprise-home-panels')
        expect(enterpriseHomePanels).toBeTruthy()
    })
})
