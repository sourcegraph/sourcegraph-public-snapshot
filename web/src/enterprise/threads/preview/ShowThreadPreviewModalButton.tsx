import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import H from 'history'
import CloseIcon from 'mdi-react/CloseIcon'
import React, { useCallback, useState } from 'react'
import { Link } from 'react-router-dom'
import { Modal } from 'reactstrap'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { PlatformContextProps } from '../../../../../shared/src/platform/context'
import { ThemeProps } from '../../../theme'
import { ThreadPreview } from './ThreadPreview'

interface Props extends ExtensionsControllerProps, PlatformContextProps, ThemeProps {
    thread: GQL.ThreadOrThreadPreview
    location: H.Location
    history: H.History
}

const REMOVE_MODAL_DIALOG_CLASS = (e: HTMLElement | null): void => {
    if (e && e.firstElementChild) {
        e.firstElementChild.classList.remove('modal-dialog')
    }
}

export const ShowThreadPreviewModalButton: React.FunctionComponent<Props> = ({
    thread,
    location,
    history,
    ...props
}) => {
    const isOpen =
        thread.__typename === 'ThreadPreview' &&
        new URLSearchParams(location.hash.slice(1)).get('thread') === thread.internalID
    const hideThreadModal = useCallback(() => {
        const hashParams = new URLSearchParams(location.hash.slice(1))
        hashParams.delete('thread')
        history.push({ ...location, hash: hashParams.toString() })
    }, [history, location])

    const [tryBuildStatus, setTryBuildStatus] = useState(false)
    const onTryBuildClick = useCallback<React.MouseEventHandler<HTMLButtonElement>>(e => {
        e.preventDefault()
        setTryBuildStatus(true)
    }, [])

    return thread.__typename === 'ThreadPreview' ? (
        <>
            <button
                type="button"
                className={`btn ${tryBuildStatus ? 'btn-link' : 'btn-secondary'} mr-2 mt-2`}
                disabled={tryBuildStatus}
                onClick={onTryBuildClick}
            >
                {tryBuildStatus ? (
                    <>
                        <LoadingSpinner className="icon-inline mr-2" /> Waiting for build
                    </>
                ) : (
                    'Try build'
                )}
            </button>
            <Link
                className="btn btn-secondary mt-2"
                to={{ ...location, hash: new URLSearchParams({ thread: thread.internalID }).toString() }}
            >
                Show preview
            </Link>
            <Modal
                isOpen={isOpen}
                backdrop={true}
                autoFocus={false}
                scrollable={true}
                fade={false}
                toggle={hideThreadModal}
                centered={true}
                className="container mx-auto"
                // The .modal-dialog class makes this modal too narrow.
                innerRef={REMOVE_MODAL_DIALOG_CLASS}
            >
                {isOpen && (
                    <div className="overflow-auto" style={{ maxHeight: '80vh' }}>
                        <ThreadPreview
                            {...props}
                            thread={thread}
                            titleRight={
                                <button type="button" className="btn btn-link p-0" aria-label="Close" onClick={hideThreadModal}>
                                    <CloseIcon className="icon-inline" /> Close preview
                                </button>
                            }
                            className="p-4"
                            location={location}
                            history={history}
                        />
                    </div>
                )}
            </Modal>
        </>
    ) : null
}
