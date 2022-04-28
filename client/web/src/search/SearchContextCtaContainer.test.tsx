import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import sinon from 'sinon'

import { AuthenticatedUser } from '@sourcegraph/shared/src/auth'
import { MockTemporarySettings } from '@sourcegraph/shared/src/settings/temporary/testUtils'
import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { renderWithBrandedContext } from '@sourcegraph/shared/src/testing'

import { useAppContext, useSearchContextCta } from '../stores'

import { SearchContextCtaContainer, SearchContextCtaContainerProps } from './SearchContextCtaContainer'

describe('SearchContextCtaContainer', () => {
    const props: SearchContextCtaContainerProps = {
        telemetryService: NOOP_TELEMETRY_SERVICE,
        isExternalServicesUserModeAll: true,
    }

    describe('with CTA', () => {
        it('should not display CTA if not on Sourcegraph.com', () => {
            useAppContext.setState({ isSourcegraphDotCom: false })

            render(
                <MockTemporarySettings settings={{ 'search.contexts.ctaDismissed': false }}>
                    <SearchContextCtaContainer {...props} />
                </MockTemporarySettings>
            )

            expect(screen.queryByRole('button', { name: /Don't show this again/ })).not.toBeInTheDocument()
        })

        it('should display CTA on Sourcegraph.com if no repos have been added and not permanently dismissed', () => {
            useAppContext.setState({ isSourcegraphDotCom: true })

            renderWithBrandedContext(
                <MockTemporarySettings settings={{ 'search.contexts.ctaDismissed': false }}>
                    <SearchContextCtaContainer {...props} />
                </MockTemporarySettings>
            )

            expect(screen.getByRole('button', { name: /Don't show this again/ })).toBeInTheDocument()
        })

        it('should not display CTA on Sourcegraph.com if user is part of an org', () => {
            const mockUserWithOrg = {
                organizations: {
                    nodes: [{ displayName: 'test org', id: '1', name: 'test' }],
                },
            } as AuthenticatedUser

            useAppContext.setState({ isSourcegraphDotCom: true, authenticatedUser: mockUserWithOrg })

            render(
                <MockTemporarySettings settings={{ 'search.contexts.ctaDismissed': false }}>
                    <SearchContextCtaContainer {...props} />
                </MockTemporarySettings>
            )

            expect(screen.queryByRole('button', { name: /Don't show this again/ })).not.toBeInTheDocument()
        })

        it('should not display CTA on Sourcegraph.com if repos have been added', () => {
            useAppContext.setState({ isSourcegraphDotCom: true })
            useSearchContextCta.setState({ hasUserAddedRepositories: true })

            render(
                <MockTemporarySettings settings={{ 'search.contexts.ctaDismissed': false }}>
                    <SearchContextCtaContainer {...props} />
                </MockTemporarySettings>
            )

            expect(screen.queryByRole('button', { name: /Don't show this again/ })).not.toBeInTheDocument()
        })

        it('should not display CTA on Sourcegraph.com if dimissed', () => {
            useAppContext.setState({ isSourcegraphDotCom: true })

            renderWithBrandedContext(
                <MockTemporarySettings settings={{ 'search.contexts.ctaDismissed': true }}>
                    <SearchContextCtaContainer {...props} />
                </MockTemporarySettings>
            )

            expect(screen.queryByRole('button', { name: /Don't show this againr/ })).not.toBeInTheDocument()
        })

        it('should dismiss CTA when clicking dismiss button', async () => {
            const onSettingsChanged = sinon.spy()
            useAppContext.setState({ isSourcegraphDotCom: true })

            renderWithBrandedContext(
                <MockTemporarySettings
                    settings={{ 'search.contexts.ctaDismissed': false }}
                    onSettingsChanged={onSettingsChanged}
                >
                    <SearchContextCtaContainer {...props} />
                </MockTemporarySettings>
            )

            // would need some time for animation before the button becomes clickable
            // otherwise we would get `unable to click element as it has or inherits pointer-events set to "none".` error
            await waitFor(() =>
                userEvent.click(screen.getByRole('button', { name: /Don't show this again/ }), undefined, {
                    skipPointerEventsCheck: true,
                })
            )

            expect(screen.queryByRole('button', { name: /Don't show this again/ })).not.toBeInTheDocument()

            sinon.assert.calledOnceWithExactly(onSettingsChanged, { 'search.contexts.ctaDismissed': true })
        })
    })
})
