import H from 'history'
import React, { useCallback } from 'react'
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
    return thread.__typename === 'ThreadPreview' ? (
        <>
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
                className="container mx-auto"
            >
                <div className="overflow-auto" style={{ maxHeight: '80vh' }}>
                    <ThreadPreview {...props} thread={thread} className="p-4" location={location} history={history} />
                </div>
            </Modal>
        </>
    ) : null
}
