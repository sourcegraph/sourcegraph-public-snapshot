import React, { useCallback, useMemo } from 'react'
import { from } from 'rxjs'
import { catchError, switchMap } from 'rxjs/operators'

import { asError, isErrorLike } from '@sourcegraph/common'
import { Button, LoadingSpinner } from '@sourcegraph/wildcard'

import { wrapRemoteObservable } from '../../api/client/api/common'
import { Link } from '../../components/Link'
import { useObservable } from '../../util/useObservable'

import { ExtensionsDevelopmentToolsProps } from '.'

export const ActiveExtensionsPanel: React.FunctionComponent<ExtensionsDevelopmentToolsProps> = props => {
    const extensionsOrError = useObservable(
        useMemo(
            () =>
                from(props.extensionsController.extHostAPI).pipe(
                    switchMap(extensionHostAPI => wrapRemoteObservable(extensionHostAPI.getActiveExtensions())),
                    catchError(error => [asError(error)])
                ),
            [props.extensionsController]
        )
    )

    const sideloadedExtensionURL = useObservable(
        useMemo(() => from(props.platformContext.sideloadedExtensionURL), [props.platformContext])
    )

    const setSideloadedExtensionURL = useCallback(() => {
        const url = window.prompt('Parcel dev server URL:', sideloadedExtensionURL || 'http://localhost:1234')
        props.platformContext.sideloadedExtensionURL.next(url)
    }, [sideloadedExtensionURL, props.platformContext])

    const clearSideloadedExtensionURL = useCallback(() => props.platformContext.sideloadedExtensionURL.next(null), [
        props.platformContext,
    ])

    return (
        <>
            <div className="card-header">Active extensions (DEBUG)</div>
            {extensionsOrError ? (
                isErrorLike(extensionsOrError) ? (
                    <div className="alert alert-danger mb-0 rounded-0">{extensionsOrError.message}</div>
                ) : extensionsOrError.length > 0 ? (
                    <div className="list-group list-group-flush">
                        {extensionsOrError.map(({ id }, index) => (
                            <div
                                key={index}
                                className="list-group-item py-2 d-flex align-items-center justify-content-between"
                            >
                                <props.link id={id} />
                            </div>
                        ))}
                    </div>
                ) : (
                    <span className="card-body">No active extensions.</span>
                )
            ) : (
                <span className="card-body">
                    <LoadingSpinner /> Loading extensions...
                </span>
            )}
            <div className="card-body border-top">
                <h4>Sideload extension</h4>
                {sideloadedExtensionURL ? (
                    <div>
                        <p>
                            <span>Load from: </span>
                            <Link to={sideloadedExtensionURL}>{sideloadedExtensionURL}</Link>
                        </p>
                        <div>
                            <Button className="mr-1" onClick={setSideloadedExtensionURL} variant="primary" size="sm">
                                Change
                            </Button>
                            <Button onClick={clearSideloadedExtensionURL} variant="danger" size="sm">
                                Clear
                            </Button>
                        </div>
                    </div>
                ) : (
                    <div>
                        <p>
                            <span>No sideloaded extension</span>
                        </p>
                        <div>
                            <Button onClick={setSideloadedExtensionURL} variant="primary" size="sm">
                                Load extension
                            </Button>
                        </div>
                    </div>
                )}
            </div>
        </>
    )
}
