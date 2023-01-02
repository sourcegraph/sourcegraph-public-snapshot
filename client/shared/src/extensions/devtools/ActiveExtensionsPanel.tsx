import React, { useMemo } from 'react'

import { from } from 'rxjs'
import { catchError, switchMap } from 'rxjs/operators'

import { asError, isErrorLike } from '@sourcegraph/common'
import { LoadingSpinner, useObservable, CardHeader, CardBody, Alert } from '@sourcegraph/wildcard'

import { wrapRemoteObservable } from '../../api/client/api/common'

import { ExtensionsDevelopmentToolsProps } from '.'

export const ActiveExtensionsPanel: React.FunctionComponent<
    React.PropsWithChildren<ExtensionsDevelopmentToolsProps>
> = props => {
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

    return (
        <>
            <CardHeader>Active extensions (DEBUG)</CardHeader>
            {extensionsOrError ? (
                isErrorLike(extensionsOrError) ? (
                    <Alert className="mb-0 rounded-0" variant="danger">
                        {extensionsOrError.message}
                    </Alert>
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
                    <CardBody>No active extensions.</CardBody>
                )
            ) : (
                <CardBody>
                    <LoadingSpinner /> Loading extensions...
                </CardBody>
            )}
        </>
    )
}
