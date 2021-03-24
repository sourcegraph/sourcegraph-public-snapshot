import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import MenuUpIcon from 'mdi-react/MenuUpIcon'
import React, { useCallback, useMemo } from 'react'
import { UncontrolledPopover } from 'reactstrap'
import { from } from 'rxjs'
import { catchError, switchMap } from 'rxjs/operators'
import { Link } from '../components/Link'
import { ExtensionsControllerProps } from './controller'
import { PlatformContextProps } from '../platform/context'
import { asError, isErrorLike } from '../util/errors'
import { useObservable } from '../util/useObservable'
import { wrapRemoteObservable } from '../api/client/api/common'

interface Props extends ExtensionsControllerProps, PlatformContextProps<'sideloadedExtensionURL'> {
    link: React.ComponentType<{ id: string }>
}

const ExtensionStatus: React.FunctionComponent<Props> = props => {
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
        <div className="extension-status card border-0">
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
                    <LoadingSpinner className="icon-inline" /> Loading extensions...
                </span>
            )}
            <div className="card-body border-top">
                <h6>Sideload extension</h6>
                {sideloadedExtensionURL ? (
                    <div>
                        <p>
                            <span>Load from: </span>
                            <Link to={sideloadedExtensionURL}>{sideloadedExtensionURL}</Link>
                        </p>
                        <div>
                            <button
                                type="button"
                                className="btn btn-sm btn-primary mr-1"
                                onClick={setSideloadedExtensionURL}
                            >
                                Change
                            </button>
                            <button
                                type="button"
                                className="btn btn-sm btn-danger"
                                onClick={clearSideloadedExtensionURL}
                            >
                                Clear
                            </button>
                        </div>
                    </div>
                ) : (
                    <div>
                        <p>
                            <span>No sideloaded extension</span>
                        </p>
                        <div>
                            <button
                                type="button"
                                className="btn btn-sm btn-primary"
                                onClick={setSideloadedExtensionURL}
                            >
                                Load extension
                            </button>
                        </div>
                    </div>
                )}
            </div>
        </div>
    )
}

/** A button that toggles the visibility of the ExtensionStatus element in a popover. */
export const ExtensionStatusPopover = React.memo<Props>(props => (
    <>
        <button type="button" id="extension-status-popover" className="btn btn-link text-decoration-none px-2">
            <span className="text-muted">Ext</span> <MenuUpIcon className="icon-inline" />
        </button>
        <UncontrolledPopover
            placement="auto-end"
            target="extension-status-popover"
            hideArrow={true}
            popperClassName="border-0"
        >
            <ExtensionStatus {...props} />
        </UncontrolledPopover>
    </>
))
