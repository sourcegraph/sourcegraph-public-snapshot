import React, { useCallback } from 'react'
import { Switch, Route, RouteComponentProps } from 'react-router'
import { from } from 'rxjs'
import { map, switchMap } from 'rxjs/operators'

import { wrapRemoteObservable } from '@sourcegraph/shared/src/api/client/api/common'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { VersionContextProps } from '@sourcegraph/shared/src/search/util'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { isErrorLike } from '@sourcegraph/shared/src/util/errors'

import { CaseSensitivityProps, PatternTypeProps, CopyQueryButtonProps, SearchContextProps } from '../search'

import { ViewPage } from './ViewPage'

interface Props
    extends RouteComponentProps<{}>,
        ExtensionsControllerProps,
        SettingsCascadeProps,
        PatternTypeProps,
        CaseSensitivityProps,
        CopyQueryButtonProps,
        VersionContextProps,
        Pick<SearchContextProps, 'selectedSearchContextSpec'>,
        TelemetryProps {
    globbing: boolean
}

/**
 * The area that handles /views routes, displaying the requested view (contributed by an extension)
 * if it exists.
 */
export const ViewsArea: React.FunctionComponent<Props> = ({ match, ...outerProps }) => {
    const getViewForID = useCallback(
        (id: string, queryParameters: Record<string, string>) =>
            from(outerProps.extensionsController.extHostAPI).pipe(
                switchMap(extensionHostAPI =>
                    wrapRemoteObservable(extensionHostAPI.getGlobalPageViews(queryParameters))
                ),
                map(views => {
                    const viewByID = views.find(view => view.id === id)
                    if (!viewByID || isErrorLike(viewByID.view)) {
                        return null
                    }
                    return viewByID.view
                })
            ),
        [outerProps.extensionsController]
    )

    return (
        <div className="container mt-4">
            {/* eslint-disable react/jsx-no-bind */}
            <Switch>
                <Route path={match.url} exact={true}>
                    <div className="alert alert-info">No view specified in the URL.</div>
                </Route>
                <Route
                    path={`${match.url}/:view`}
                    render={({ match, location, ...props }: RouteComponentProps<{ view: string }>) => (
                        <ViewPage
                            {...outerProps}
                            {...props}
                            getViewForID={getViewForID}
                            viewID={match.params.view}
                            extraPath={location.pathname.slice(match.url.length)}
                            location={location}
                        />
                    )}
                />
            </Switch>
            {/* eslint-enable react/jsx-no-bind */}
        </div>
    )
}
