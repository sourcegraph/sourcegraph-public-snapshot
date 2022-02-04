import classNames from 'classnames'
import React, { useMemo } from 'react'
import { BrowserRouter, Route, RouteComponentProps, Switch } from 'react-router-dom'

import { createController as createExtensionsController } from '@sourcegraph/shared/src/extensions/controller'
import { aggregateStreamingSearch } from '@sourcegraph/shared/src/search/stream'
import { EMPTY_SETTINGS_CASCADE } from '@sourcegraph/shared/src/settings/settings'
import { isMacPlatform } from '@sourcegraph/shared/src/util/browserDetection'
import { AnchorLink, setLinkComponent } from '@sourcegraph/wildcard'

import { createPlatformContext } from '../../platform/context'
import { fetchHighlightedFileLineRanges, fetchRepository, resolveRevision } from '../../repo/backend'
import '../../SourcegraphWebApp.scss'
import { EmbeddedNotebookPage } from '../../search/notebook/EmbeddedNotebookPage'
import { eventLogger } from '../../tracking/eventLogger'

setLinkComponent(AnchorLink)

export const EmbeddedWebApp: React.FunctionComponent = () => {
    const platformContext = useMemo(() => createPlatformContext(), [])
    const extensionsController = useMemo(() => createExtensionsController(platformContext), [platformContext])
    const isLightTheme = true

    return (
        <BrowserRouter>
            <div className={classNames(isLightTheme ? 'theme-light' : 'theme-dark', 'p-3')}>
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
                                isMacPlatform={isMacPlatform}
                                resolveRevision={resolveRevision}
                                fetchRepository={fetchRepository}
                                streamSearch={aggregateStreamingSearch}
                                platformContext={platformContext}
                                extensionsController={extensionsController}
                                settingsCascade={EMPTY_SETTINGS_CASCADE}
                            />
                        )}
                    />
                </Switch>
            </div>
        </BrowserRouter>
    )
}
