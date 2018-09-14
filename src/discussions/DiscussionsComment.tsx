import LinkIcon from '@sourcegraph/icons/lib/Link'
import copy from 'copy-to-clipboard'
import * as H from 'history'
import * as React from 'react'
import { Link } from 'react-router-dom'
import * as GQL from '../backend/graphqlschema'
import { Markdown } from '../components/Markdown'
import { Timestamp } from '../components/time/Timestamp'
import { eventLogger } from '../tracking/eventLogger'
import { UserAvatar } from '../user/UserAvatar'

interface Props {
    comment: GQL.IDiscussionComment
    threadID: GQL.ID
    location: H.Location
}

interface State {
    copiedLink: boolean
}

export class DiscussionsComment extends React.PureComponent<Props> {
    private scrollToElement: HTMLElement | null = null

    public state: State = {
        copiedLink: false,
    }

    public componentDidMount(): void {
        if (this.scrollToElement) {
            this.scrollToElement.scrollIntoView()
        }
    }

    public render(): JSX.Element | null {
        const { location, comment } = this.props
        const isTargeted = new URLSearchParams(location.hash).get('commentID') === comment.id

        // TODO(slimsag:discussions): ASAP: markdown links, headings, etc lead to #

        return (
            <div
                className={`discussions-comment${isTargeted ? ' discussions-comment--targeted' : ''}`}
                ref={isTargeted ? this.setScrollToElement : undefined}
            >
                <div className="discussions-comment__top-area">
                    <span className="discussions-comment__author">
                        <Link to={`/users/${comment.author.username}`} data-tooltip={comment.author.displayName}>
                            <UserAvatar user={comment.author} className="icon-inline icon-sm" />
                        </Link>
                        <Link
                            to={`/users/${comment.author.username}`}
                            data-tooltip={comment.author.displayName}
                            className="discussions-comment__author-name"
                        >
                            {comment.author.username}
                        </Link>
                        {' commented'}
                    </span>
                    <span className="discussions-comment__spacer" />
                    <span className="discussions-comment__top-right-area">
                        {/* TODO(slimsag:discussions): timestamp should not wrap around on small screen widths */}
                        <Timestamp date={comment.createdAt} />
                        {this.props.comment.inlineURL && (
                            <Link
                                className="btn btn-link btn-sm discussions-comment__share"
                                data-tooltip="Copy link to this comment"
                                to={this.props.comment.inlineURL}
                                onClick={this.onShareLinkClick}
                            >
                                {this.state.copiedLink ? 'Copied!' : <LinkIcon className="icon-inline" />}
                            </Link>
                        )}
                    </span>
                </div>
                <div className="discussions-comment__content">
                    <Markdown dangerousInnerHTML={comment.html} />
                </div>
            </div>
        )
    }

    private onShareLinkClick: React.MouseEventHandler<HTMLElement> = event => {
        if (event.metaKey || event.altKey || event.ctrlKey) {
            return
        }
        eventLogger.log('ShareCommentButtonClicked')
        event.preventDefault()
        copy(this.props.comment.inlineURL!) // ! because this method is only called when inlineURL exists
        this.setState({ copiedLink: true })
        setTimeout(() => {
            this.setState({ copiedLink: false })
        }, 1000)
    }

    private setScrollToElement = (ref: HTMLElement | null) => {
        this.scrollToElement = ref
    }
}
