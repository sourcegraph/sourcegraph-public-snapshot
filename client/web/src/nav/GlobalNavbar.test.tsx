import { render } from '@testing-library/react'
import { createLocation, createMemoryHistory } from 'history'
import React from 'react'
import { MemoryRouter } from 'react-router'

import { setLinkComponent } from '@sourcegraph/shared/src/components/Link'
import { extensionsController, NOOP_SETTINGS_CASCADE } from '@sourcegraph/shared/src/testing/searchTestHelpers'

import { useExperimentalFeatures, useNavbarQueryState } from '../stores'
import { ThemePreference } from '../stores/themeState'

import { GlobalNavbar } from './GlobalNavbar'

jest.mock('../search/input/SearchNavbarItem', () => ({ SearchNavbarItem: 'SearchNavbarItem' }))
jest.mock('../components/branding/BrandLogo', () => ({ BrandLogo: 'BrandLogo' }))

const PROPS: React.ComponentProps<typeof GlobalNavbar> = {
    authenticatedUser: null,
    authRequired: false,
    extensionsController,
    location: createLocation('/'),
    history: createMemoryHistory(),
    keyboardShortcuts: [],
    isSourcegraphDotCom: false,
    onThemePreferenceChange: () => undefined,
    isLightTheme: true,
    themePreference: ThemePreference.Light,
    platformContext: {} as any,
    settingsCascade: NOOP_SETTINGS_CASCADE,
    batchChangesEnabled: false,
    batchChangesExecutionEnabled: false,
    batchChangesWebhookLogsEnabled: false,
    telemetryService: {} as any,
    isExtensionAlertAnimating: false,
    showSearchBox: true,
    variant: 'default',
    globbing: false,
    branding: undefined,
    routes: [],
    extensionViews: () => null,
}

describe('GlobalNavbar', () => {
    setLinkComponent(({ children, ...props }) => <a {...props}>{children}</a>)
    afterAll(() => setLinkComponent(() => null)) // reset global env for other tests
    beforeEach(() => {
        useExperimentalFeatures.setState({ codeMonitoring: false, showSearchContext: true })
        useNavbarQueryState.setState({ searchContextsEnabled: true })
    })

    test('default', () => {
        const { asFragment } = render(
            <MemoryRouter>
                <GlobalNavbar {...PROPS} />
            </MemoryRouter>
        )
        expect(asFragment()).toMatchSnapshot()
    })

    test('low-profile', () => {
        const { asFragment } = render(
            <MemoryRouter>
                <GlobalNavbar {...PROPS} variant="low-profile" />
            </MemoryRouter>
        )
        expect(asFragment()).toMatchSnapshot()
    })
})
