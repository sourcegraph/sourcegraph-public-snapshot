import { render, fireEvent } from '@testing-library/react'
import { mount } from 'enzyme'
import React from 'react'
import { act } from 'react-dom/test-utils'
import { Dropdown, DropdownToggle } from 'reactstrap'
import sinon from 'sinon'

import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'
import {
    mockFetchAutoDefinedSearchContexts,
    mockFetchSearchContexts,
    mockGetUserSearchContextNamespaces,
} from '@sourcegraph/shared/src/testing/searchContexts/testHelpers'
import { MockIntersectionObserver } from '@sourcegraph/shared/src/util/MockIntersectionObserver'

<<<<<<< HEAD
import { SearchPatternType } from '../../graphql-operations'
import { MockTemporarySettings } from '../../settings/temporary/testUtils'

=======
>>>>>>> main
import { SearchContextDropdown, SearchContextDropdownProps } from './SearchContextDropdown'
import { SearchContextMenuItem } from './SearchContextMenu'

describe('SearchContextDropdown', () => {
    const defaultProps: SearchContextDropdownProps = {
        telemetryService: NOOP_TELEMETRY_SERVICE,
        query: '',
        showSearchContextManagement: false,
        fetchAutoDefinedSearchContexts: mockFetchAutoDefinedSearchContexts(1),
        fetchSearchContexts: mockFetchSearchContexts,
        getUserSearchContextNamespaces: mockGetUserSearchContextNamespaces,
        defaultSearchContextSpec: '',
        selectedSearchContextSpec: '',
        setSelectedSearchContextSpec: () => {},
        hasUserAddedRepositories: false,
        hasUserAddedExternalServices: false,
        isSourcegraphDotCom: false,
        authenticatedUser: null,
        searchContextsEnabled: true,
    }
    const RealIntersectionObserver = window.IntersectionObserver
    let clock: sinon.SinonFakeTimers

    beforeAll(() => {
        clock = sinon.useFakeTimers()
        window.IntersectionObserver = MockIntersectionObserver
    })

    afterAll(() => {
        clock.restore()
        window.IntersectionObserver = RealIntersectionObserver
    })

    it('should start closed', () => {
        const element = mount(<SearchContextDropdown {...defaultProps} />)
        const button = element.find(Dropdown)
        expect(button.prop('isOpen')).toBe(false)
    })

    it('should open when toggle event happens', () => {
        const element = mount(<SearchContextDropdown {...defaultProps} />)
        let button = element.find(Dropdown)
        button.invoke('toggle')?.(new MouseEvent('click') as any)

        button = element.find(Dropdown)
        expect(button.prop('isOpen')).toBe(true)
    })

    it('should close if toggle event happens again', () => {
        const element = mount(<SearchContextDropdown {...defaultProps} />)
        let button = element.find(Dropdown)
        button.invoke('toggle')?.(new MouseEvent('click') as any)

        button = element.find(Dropdown)
        button.invoke('toggle')?.(new MouseEvent('click') as any)

        button = element.find(Dropdown)
        expect(button.prop('isOpen')).toBe(false)
    })

    it('should be enabled if query is empty', () => {
        const element = mount(<SearchContextDropdown {...defaultProps} />)
        const dropdown = element.find(DropdownToggle)
        expect(dropdown.prop('disabled')).toBe(false)
        expect(dropdown.prop('data-tooltip')).toBe('')
    })

    it('should be enabled if query does not contain context filter', () => {
        const element = mount(<SearchContextDropdown {...defaultProps} query="test (repo:foo or repo:python)" />)
        const dropdown = element.find(DropdownToggle)
        expect(dropdown.prop('disabled')).toBe(false)
        expect(dropdown.prop('data-tooltip')).toBe('')
    })

    it('should be disabled if query contains context filter', () => {
        const element = mount(<SearchContextDropdown {...defaultProps} query="test (context:foo or repo:python)" />)
        const dropdown = element.find(DropdownToggle)
        expect(dropdown.prop('disabled')).toBe(true)
        expect(dropdown.prop('data-tooltip')).toBe('Overridden by query')
    })

    it('should submit search on item click', () => {
        const submitSearch = sinon.spy()
        const element = mount(<SearchContextDropdown {...defaultProps} submitSearch={submitSearch} query="test" />)

        act(() => {
            // Wait for debounce
            clock.tick(50)
        })
        element.update()

        const item = element.find(SearchContextMenuItem).at(0)
        item.simulate('click')

        sinon.assert.calledOnce(submitSearch)
    })
<<<<<<< HEAD

    it('should not submit search if submitSearchOnSearchContextChange is false', () => {
        const submitSearch = sinon.spy()
        const element = mount(
            <SearchContextDropdown
                {...defaultProps}
                submitSearch={submitSearch}
                submitSearchOnSearchContextChange={false}
            />
        )

        act(() => {
            // Wait for debounce
            clock.tick(50)
        })
        element.update()

        const item = element.find(DropdownItem).at(0)
        item.simulate('click')

        sinon.assert.notCalled(submitSearch)
    })

    describe('with CTA', () => {
        let oldContext: any
        beforeEach(() => {
            oldContext = window.context
            window.context = { externalServicesUserMode: 'all' } as any
        })

        afterEach(() => {
            window.context = oldContext
        })

        it('should display CTA on Sourcegraph.com if no repos have been added and not permanently dismissed', () => {
            const { getByRole, queryByRole } = render(
                <MockTemporarySettings settings={{ 'search.contexts.ctaPermanentlyDismissed': false }}>
                    <SearchContextDropdown
                        {...defaultProps}
                        isSourcegraphDotCom={true}
                        hasUserAddedRepositories={false}
                    />
                </MockTemporarySettings>
            )

            fireEvent.click(getByRole('button', { name: /context:/ }))

            expect(queryByRole('button', { name: /Maybe later/ })).toBeInTheDocument()
        })

        it('should not display CTA on Sourcegraph.com if repos have been added', () => {
            const { getByRole, queryByRole } = render(
                <MockTemporarySettings settings={{ 'search.contexts.ctaPermanentlyDismissed': false }}>
                    <SearchContextDropdown
                        {...defaultProps}
                        isSourcegraphDotCom={true}
                        hasUserAddedRepositories={true}
                    />
                </MockTemporarySettings>
            )

            fireEvent.click(getByRole('button', { name: /context:/ }))

            expect(queryByRole('button', { name: /Maybe later/ })).not.toBeInTheDocument()
        })

        it('should not display CTA on Sourcegraph.com if permnanently dimissed', () => {
            const { getByRole, queryByRole } = render(
                <MockTemporarySettings settings={{ 'search.contexts.ctaPermanentlyDismissed': true }}>
                    <SearchContextDropdown
                        {...defaultProps}
                        isSourcegraphDotCom={true}
                        hasUserAddedRepositories={false}
                    />
                </MockTemporarySettings>
            )

            fireEvent.click(getByRole('button', { name: /context:/ }))

            expect(queryByRole('button', { name: /Maybe later/ })).not.toBeInTheDocument()
        })

        it('should should dismiss CTA (but not permanetly) when clicking dismiss button', () => {
            const onSettingsChanged = sinon.spy()

            const { getByRole, queryByRole } = render(
                <MockTemporarySettings
                    settings={{ 'search.contexts.ctaPermanentlyDismissed': false }}
                    onSettingsChanged={onSettingsChanged}
                >
                    <SearchContextDropdown
                        {...defaultProps}
                        isSourcegraphDotCom={true}
                        hasUserAddedRepositories={false}
                    />
                </MockTemporarySettings>
            )

            fireEvent.click(getByRole('button', { name: /context:/ }))
            fireEvent.click(getByRole('button', { name: /Maybe later/ }))

            expect(queryByRole('button', { name: /Maybe later/ })).not.toBeInTheDocument()
            expect(getByRole('searchbox')).toBeInTheDocument()

            sinon.assert.notCalled(onSettingsChanged)
        })

        it('should should dismiss CTA when clicking dismiss button and checkbox enabled', () => {
            const onSettingsChanged = sinon.spy()

            const { getByRole, queryByRole } = render(
                <MockTemporarySettings
                    settings={{ 'search.contexts.ctaPermanentlyDismissed': false }}
                    onSettingsChanged={onSettingsChanged}
                >
                    <SearchContextDropdown
                        {...defaultProps}
                        isSourcegraphDotCom={true}
                        hasUserAddedRepositories={false}
                    />
                </MockTemporarySettings>
            )

            fireEvent.click(getByRole('button', { name: /context:/ }))

            fireEvent.click(getByRole('checkbox', { name: /Don't show this again/ }))
            fireEvent.click(getByRole('button', { name: /Maybe later/ }))

            expect(queryByRole('button', { name: /Maybe later/ })).not.toBeInTheDocument()
            expect(getByRole('searchbox')).toBeInTheDocument()

            sinon.assert.calledWithExactly(onSettingsChanged, { 'search.contexts.ctaPermanentlyDismissed': true })
        })
    })
=======
>>>>>>> main
})
