import * as H from 'history'
import * as React from 'react'
import { Link } from 'react-router-dom'
import { ChatIcon } from '../../../../../shared/src/components/icons'
import { LineOrPositionOrRange, RepoFile } from '../../../../../shared/src/util/url'
import { eventLogger } from '../../../tracking/eventLogger'

interface DiscussionsGutterOverlayProps extends RepoFile {
    location: H.Location

    /** The currently selected position. */
    selectedPosition: LineOrPositionOrRange

    /** The position of the tooltip (assigned to `style`) */
    overlayPosition?: { left: number; top: number }
}

const onCreateDiscussionClick = (): void => eventLogger.log('CreateDiscussionClicked')

export const DiscussionsGutterOverlay: React.FunctionComponent<DiscussionsGutterOverlayProps> = props => {
    const hash = new URLSearchParams(props.location.hash.slice('#'.length))
    const onDiscussionsNew = hash.get('tab') === 'discussions' && hash.get('createThread') === 'true'
    hash.delete('threadID')
    hash.delete('commentID')
    if (onDiscussionsNew) {
        hash.delete('tab')
        hash.delete('createThread')
    } else {
        hash.set('tab', 'discussions')
        hash.set('createThread', 'true')
    }
    const newURL = props.location.pathname + props.location.search + '#' + hash.toString()

    return (
        <div
            className="discussions-gutter-overlay"
            // needed for dynamic styling
            // eslint-disable-next-line react/forbid-dom-props
            style={
                props.overlayPosition
                    ? {
                          position: 'absolute',
                          opacity: 1,
                          visibility: 'visible',
                          left: props.overlayPosition.left + 'px',
                          top: props.overlayPosition.top + 'px',
                      }
                    : {
                          opacity: 0,
                          visibility: 'hidden',
                      }
            }
        >
            <Link
                className="discussions-gutter-overlay__link btn btn-sm btn-link btn-icon"
                onClick={onCreateDiscussionClick}
                data-tooltip={onDiscussionsNew ? 'Close discussions' : 'Create a discussion for this selection'}
                to={newURL}
            >
                <ChatIcon className="icon-inline" />
            </Link>
        </div>
    )
}
