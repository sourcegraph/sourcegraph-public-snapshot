import { FC, Suspense, useEffect, useMemo } from 'react'

import { Router } from 'react-router'
import { CompatRouter, Routes, Route } from 'react-router-dom-v5-compat'

import { createController as createExtensionsController } from '@sourcegraph/shared/src/extensions/createSyncLoadedController'
import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'
import { Alert, LoadingSpinner, setLinkComponent, WildcardTheme, WildcardThemeContext } from '@sourcegraph/wildcard'

import '../../SourcegraphWebApp.scss'

import { GlobalContributions } from '../../contributions'
import { createPlatformContext } from '../../platform/context'
import { useTheme, ThemePreference } from '../../theme'
import { globalHistory } from '../../util/globalHistory'

import { OpenNewTabAnchorLink } from './OpenNewTabAnchorLink'

import styles from './EmbeddedWebApp.module.scss'

// Since we intend to embed the EmbeddedWebApp component within an iframe,
// we want to open all links in a new tab instead of the current iframe window.
// Otherwise, we would get an error that we tried to access a non-embed route from within the iframe.
setLinkComponent(OpenNewTabAnchorLink)

const WILDCARD_THEME: WildcardTheme = {
    isBranded: true,
}

const EmbeddedNotebookPage = lazyComponent(
    () => import('../../notebooks/notebookPage/EmbeddedNotebookPage'),
    'EmbeddedNotebookPage'
)

const EMPTY_SETTINGS_CASCADE = { final: {}, subjects: [] }

export const EmbeddedWebApp: FC = () => {
    const { enhancedThemePreference, setThemePreference } = useTheme()
    const isLightTheme = enhancedThemePreference === ThemePreference.Light

    useEffect(() => {
        const query = new URLSearchParams(window.location.search)
        const theme = query.get('theme')
        setThemePreference(
            theme === 'dark' ? ThemePreference.Dark : theme === 'light' ? ThemePreference.Light : ThemePreference.System
        )
    }, [setThemePreference])

    useEffect(() => {
        document.documentElement.classList.toggle('theme-light', isLightTheme)
        document.documentElement.classList.toggle('theme-dark', !isLightTheme)
    }, [isLightTheme])

    const platformContext = useMemo(() => createPlatformContext(), [])
    const extensionsController = useMemo(() => createExtensionsController(platformContext), [platformContext])

    // 🚨 SECURITY: The `EmbeddedWebApp` is intended to be embedded into 3rd party sites where we do not have total control.
    // That is why it is essential to be mindful when adding new routes that may be vulnerable to clickjacking or similar exploits.
    // It is crucial not to embed any components that an attacker could hijack and use to leak personal information (e.g., the sign-in page).
    // The embedded components should be limited to displaying read-only, publicly accessible content.
    //
    // IMPORTANT: Please consult with the security team if you are unsure whether your changes could introduce security exploits.
    return (
        <Router history={globalHistory}>
            <CompatRouter>
                <WildcardThemeContext.Provider value={WILDCARD_THEME}>
                    <div className={styles.body}>
                        <Suspense
                            fallback={
                                <div className="d-flex justify-content-center p-3">
                                    <LoadingSpinner />
                                </div>
                            }
                        >
                            <Routes>
                                <Route
                                    path="/embed/notebooks/:notebookId"
                                    element={
                                        <EmbeddedNotebookPage
                                            searchContextsEnabled={true}
                                            isSourcegraphDotCom={window.context.sourcegraphDotComMode}
                                            authenticatedUser={null}
                                            isLightTheme={isLightTheme}
                                            settingsCascade={EMPTY_SETTINGS_CASCADE}
                                            platformContext={platformContext}
                                        />
                                    }
                                />
                                <Route
                                    path="*"
                                    element={
                                        <Alert variant="danger">
                                            Invalid embedding route, please check the embedding URL.
                                        </Alert>
                                    }
                                />
                            </Routes>
                        </Suspense>
                        <GlobalContributions
                            extensionsController={extensionsController}
                            platformContext={platformContext}
                        />
                    </div>
                </WildcardThemeContext.Provider>
            </CompatRouter>
        </Router>
    )
}
