import React, { useCallback, useMemo } from 'react'

import { from } from 'rxjs'
import { catchError, switchMap } from 'rxjs/operators'

import { asError, isErrorLike } from '@sourcegraph/common'
import {
    Button,
    LoadingSpinner,
    useObservable,
    Link,
    CardHeader,
    CardBody,
    Alert,
    H4,
    Text,
} from '@sourcegraph/wildcard'

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

    const sideloadedExtensionURL = useObservable(
        useMemo(() => from(props.platformContext.sideloadedExtensionURL), [props.platformContext])
    )

    const setSideloadedExtensionURL = useCallback(() => {
        const url = window.prompt('Parcel dev server URL:', sideloadedExtensionURL || 'http://localhost:1234')
        props.platformContext.sideloadedExtensionURL.next(url)
    }, [sideloadedExtensionURL, props.platformContext])

    const clearSideloadedExtensionURL = useCallback(
        () => props.platformContext.sideloadedExtensionURL.next(null),
        [props.platformContext]
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
            <CardBody className="border-top">
                <H4>Sideload extension</H4>
                {sideloadedExtensionURL ? (
                    <div>
                        <Text>
                            <span>Load from: </span>
                            <Link to={sideloadedExtensionURL}>{sideloadedExtensionURL}</Link>
                        </Text>
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
                        <Text>
                            <span>No sideloaded extension</span>
                        </Text>
                        <div>
                            <Button onClick={setSideloadedExtensionURL} variant="primary" size="sm">
                                Load extension
                            </Button>
                        </div>
                    </div>
                )}
            </CardBody>
        </>
    )
}
