import classNames from 'classnames'
import React, { Suspense, useMemo } from 'react'
import { BrowserRouter, Route, RouteComponentProps, Switch } from 'react-router-dom'

import { isMacPlatform } from '@sourcegraph/common'
import { createController as createExtensionsController } from '@sourcegraph/shared/src/extensions/controller'
import { aggregateStreamingSearch } from '@sourcegraph/shared/src/search/stream'
import { EMPTY_SETTINGS_CASCADE } from '@sourcegraph/shared/src/settings/settings'
import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'
import {
    Alert,
    AnchorLink,
    LoadingSpinner,
    setLinkComponent,
    WildcardTheme,
    WildcardThemeContext,
} from '@sourcegraph/wildcard'

import { createPlatformContext } from '../../platform/context'
import { fetchHighlightedFileLineRanges, fetchRepository, resolveRevision } from '../../repo/backend'
import '../../SourcegraphWebApp.scss'
import { eventLogger } from '../../tracking/eventLogger'

setLinkComponent(AnchorLink)

const WILDCARD_THEME: WildcardTheme = {
    isBranded: true,
}

const EmbeddedNotebookPage = lazyComponent(
    () => import('../../search/notebook/EmbeddedNotebookPage'),
    'EmbeddedNotebookPage'
)

export const EmbeddedWebApp: React.FunctionComponent = () => {
    const platformContext = useMemo(() => createPlatformContext(), [])
    const extensionsController = useMemo(() => createExtensionsController(platformContext), [platformContext])
    // We only support light theme for now, but this can be made dynamic through a URL param in the embedding link.
    const isLightTheme = true

    // 🚨 SECURITY: The `EmbeddedWebApp` is intended to be embedded into 3rd party sites where we do not have total control.
    // That is why it is essential to be mindful when adding new routes that may be vulnerable to clickjacking or similar exploits.
    // It is crucial not to embed any components that an attacker could hijack and use to leak personal information (e.g., the sign-in page).
    // The embedded components should be limited to displaying read-only, publicly accessible content.
    //
    // IMPORTANT: Please consult with the security team if you are unsure whether your changes could introduce security exploits.
    return (
        <BrowserRouter>
            <WildcardThemeContext.Provider value={WILDCARD_THEME}>
                <div className={classNames(isLightTheme ? 'theme-light' : 'theme-dark', 'p-3')}>
                    <Suspense
                        fallback={
                            <div className="d-flex justify-content-center">
                                <LoadingSpinner />
                            </div>
                        }
                    >
                        <Switch>
                            <Route
                                path="/embed/notebooks/:notebookId"
                                render={(props: RouteComponentProps<{ notebookId: string }>) => (
                                    <EmbeddedNotebookPage
                                        notebookId={props.match.params.notebookId}
                                        searchContextsEnabled={false}
                                        showSearchContext={false}
                                        isSourcegraphDotCom={window.context.sourcegraphDotComMode}
                                        authenticatedUser={null}
                                        fetchHighlightedFileLineRanges={fetchHighlightedFileLineRanges}
                                        isLightTheme={isLightTheme}
                                        telemetryService={eventLogger}
                                        globbing={true}
                                        isMacPlatform={isMacPlatform()}
                                        resolveRevision={resolveRevision}
                                        fetchRepository={fetchRepository}
                                        streamSearch={aggregateStreamingSearch}
                                        platformContext={platformContext}
                                        extensionsController={extensionsController}
                                        settingsCascade={EMPTY_SETTINGS_CASCADE}
                                    />
                                )}
                            />
                            <Route
                                path="*"
                                render={() => (
                                    <Alert variant="danger">
                                        Invalid embedding route, please check the embedding URL.
                                    </Alert>
                                )}
                            />
                        </Switch>
                    </Suspense>
                </div>
            </WildcardThemeContext.Provider>
        </BrowserRouter>
    )
}
