import React, { useMemo } from 'react'
import { ExtensionsControllerProps } from '../../../shared/src/extensions/controller'
import { PageTitle } from '../components/PageTitle'
import * as H from 'history'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { useObservable } from '../../../shared/src/util/useObservable'
import { ViewContentProps, ViewContent } from './ViewContent'
import { map, switchMap } from 'rxjs/operators'
import { from } from 'rxjs'
import { wrapRemoteObservable } from '../../../shared/src/api/client/api/common'
import { isErrorLike } from '../../../shared/src/util/errors'

interface Props
    extends ExtensionsControllerProps<'extHostAPI'>,
        Omit<ViewContentProps, 'viewContent' | 'containerClassName'> {
    viewID: string
    extraPath: string

    location: H.Location
    history: H.History
}

/**
 * A page that displays a single view (contributed by an extension) as a standalone page.
 */
export const ViewPage: React.FunctionComponent<Props> = ({
    viewID,
    extraPath,
    location,
    extensionsController,
    ...props
}) => {
    const queryParameters = useMemo<Record<string, string>>(
        () => ({ ...Object.fromEntries(new URLSearchParams(location.search).entries()), extraPath }),
        [extraPath, location.search]
    )

    const contributions = useMemo(
        () =>
            from(extensionsController.extHostAPI).pipe(
                switchMap(extensionHostAPI => wrapRemoteObservable(extensionHostAPI.getContributions()))
            ),
        [extensionsController]
    )

    const view = useObservable(
        useMemo(
            () =>
                // TODO: implement getViewByID (not urgent)
                from(extensionsController.extHostAPI).pipe(
                    switchMap(extensionHostAPI =>
                        wrapRemoteObservable(extensionHostAPI.getGlobalPageViews(queryParameters))
                    ),
                    map(views => {
                        const viewByID = views.find(view => view.id === viewID)
                        if (!viewByID || isErrorLike(viewByID.view)) {
                            return null
                        }
                        return viewByID.view
                    })
                ),
            [contributions, extensionsController, queryParameters, viewID]
        )
    )

    if (view === undefined) {
        return <LoadingSpinner className="icon-inline" />
    }

    if (view === null) {
        return (
            <div className="alert alert-danger">
                View not found: <code>{viewID}</code>
            </div>
        )
    }

    return (
        <div>
            <PageTitle title={view.title || 'View'} />
            {view.title && <h1>{view.title}</h1>}
            <ViewContent viewID={viewID} viewContent={view.content} location={location} {...props} />
        </div>
    )
}
