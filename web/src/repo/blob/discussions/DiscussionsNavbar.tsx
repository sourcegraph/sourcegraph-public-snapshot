import * as H from 'history'
import ChevronRightIcon from 'mdi-react/ChevronRightIcon'
import * as React from 'react'
import { Link } from 'react-router-dom'
import * as GQL from '../../../../../shared/src/graphql/schema'

interface Props {
    threadID?: GQL.ID
    threadTitle?: string
    commentID?: GQL.ID
    commentContent?: string
    filePath?: string
    location: H.Location
}

export class DiscussionsNavbar extends React.PureComponent<Props> {
    public render(): JSX.Element | null {
        // TODO(slimsag:discussions): make ID number smaller and grey like on thread list
        const { threadID, threadTitle, commentID, commentContent, filePath, location } = this.props
        return (
            <div className="discussions-navbar">
                {filePath && <Link to={this.locationWith(location)}>{filePath}</Link>}
                <ChevronRightIcon className="icon-inline" />
                {threadID !== undefined && commentID !== undefined && (
                    <>
                        <Link to={this.locationWith(location, threadID)}>
                            {threadTitle !== undefined && `${threadTitle} `}#{threadID}
                        </Link>
                        <ChevronRightIcon className="icon-inline" />
                        <strong>
                            {commentContent !== undefined && `${commentContent} `}#{this.props.commentID}
                        </strong>
                    </>
                )}
                {threadID !== undefined && commentID === undefined && (
                    <strong>
                        {threadTitle !== undefined && `${threadTitle} `}#{threadID}
                    </strong>
                )}
                {threadID === undefined && commentID === undefined && (
                    <>
                        {!threadTitle && <strong>New discussion</strong>}
                        {threadTitle && (
                            <>
                                <strong className="discussions-navbar__title-container">
                                    New discussion:{' '}
                                    <small className="discussions-navbar__title text-muted">"{threadTitle}"</small>
                                </strong>
                            </>
                        )}
                    </>
                )}
            </div>
        )
    }

    private locationWith(location: H.Location, threadID?: GQL.ID): string {
        // TODO(slimsag:discussions): future: for correctness, this should not
        // assume the current location and instead use this.props.filePath etc.
        const hash = new URLSearchParams(location.hash.slice('#'.length))
        hash.set('tab', 'discussions')
        hash.delete('createThread')
        hash.delete('commentID')
        if (threadID) {
            hash.set('threadID', threadID)
        } else {
            hash.delete('threadID')
        }
        return location.pathname + location.search + '#' + hash.toString()
    }
}
