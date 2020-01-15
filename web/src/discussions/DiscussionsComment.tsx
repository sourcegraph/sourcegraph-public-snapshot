import copy from 'copy-to-clipboard'
import * as H from 'history'
import CommentCheckIcon from 'mdi-react/CommentCheckIcon'
import CommentRemoveIcon from 'mdi-react/CommentRemoveIcon'
import FlagVariantIcon from 'mdi-react/FlagVariantIcon'
import LinkIcon from 'mdi-react/LinkIcon'
import * as React from 'react'
import { Link } from 'react-router-dom'
import { Observable } from 'rxjs'
import { WithLinkPreviews } from '../../../shared/src/components/linkPreviews/WithLinkPreviews'
import { Markdown } from '../../../shared/src/components/Markdown'
import { ExtensionsControllerProps } from '../../../shared/src/extensions/controller'
import * as GQL from '../../../shared/src/graphql/schema'
import { asError } from '../../../shared/src/util/errors'
import { LINK_PREVIEW_CLASS } from '../components/linkPreviews/styles'
import { Timestamp } from '../components/time/Timestamp'
import { setElementTooltip } from '../components/tooltip/Tooltip'
import { eventLogger } from '../tracking/eventLogger'
import { UserAvatar } from '../user/UserAvatar'

interface Props extends ExtensionsControllerProps {
    comment: GQL.IDiscussionComment
    threadID: GQL.ID
    location: H.Location

    /**
     * When specified, a report icon will be displayed inline and that function
     * will be called when a report has been submitted.
     */
    onReport?: (comment: GQL.IDiscussionComment, reason: string) => Observable<void>

    /**
     * When specified, that function is called to handle the
     * "Clear reports / mark as read" button clicks.
     */
    onClearReports?: (comment: GQL.IDiscussionComment) => Observable<void>

    /**
     * When specified, that function is called to handle the "delete comment"
     * button clicks.
     */
    onDelete?: (comment: GQL.IDiscussionComment) => Observable<void>
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
        if (that.scrollToElement) {
            that.scrollToElement.scrollIntoView()
        }
    }

    public render(): JSX.Element | null {
        const { location, comment, onReport, onClearReports, onDelete } = that.props
        const isTargeted = new URLSearchParams(location.hash).get('commentID') === comment.idWithoutKind

        // TODO(slimsag:discussions): ASAP: markdown links, headings, etc lead to #

        return (
            <div
                className={`discussions-comment${isTargeted ? ' discussions-comment--targeted' : ''}`}
                ref={isTargeted ? that.setScrollToElement : undefined}
            >
                <div className="discussions-comment__top-area">
                    <span className="discussions-comment__author">
                        <Link to={`/users/${comment.author.username}`} data-tooltip={comment.author.displayName}>
                            <UserAvatar user={comment.author} className="icon-inline icon-sm" />
                        </Link>
                        <Link
                            to={`/users/${comment.author.username}`}
                            data-tooltip={comment.author.displayName}
                            className="ml-1 mr-1"
                        >
                            {comment.author.username}
                        </Link>
                        <span className="mr-1">commented</span>
                        <Timestamp date={comment.createdAt} />
                    </span>
                    <span className="discussions-comment__spacer" />
                    <span className="discussions-comment__top-right-area">
                        {that.props.comment.inlineURL && (
                            <Link
                                className="btn btn-link btn-sm discussions-comment__share"
                                data-tooltip="Copy link to this comment"
                                to={that.props.comment.inlineURL}
                                onClick={that.onShareLinkClick}
                            >
                                {that.state.copiedLink ? 'Copied!' : <LinkIcon className="icon-inline" />}
                            </Link>
                        )}

                        {comment.canReport && onReport && (
                            <button
                                type="button"
                                className="btn btn-link btn-sm discussions-comment__report"
                                data-tooltip="Report this comment"
                                onClick={that.onReportClick}
                            >
                                <FlagVariantIcon className="icon-inline" />
                            </button>
                        )}
                        {comment.reports.length > 0 && (
                            <>
                                <span
                                    className="ml-1 mr-1 discussions-comment__reports"
                                    data-tooltip={comment.reports.join('\n\n')}
                                >
                                    {comment.reports.length} reports
                                </span>
                                {comment.canClearReports && onClearReports && (
                                    <button
                                        type="button"
                                        className="btn btn-link btn-sm discussions-comment__toolbar-btn"
                                        data-tooltip="Clear reports / mark as good message"
                                        onClick={that.onClearReportsClick}
                                    >
                                        <CommentCheckIcon className="icon-inline" />
                                    </button>
                                )}
                            </>
                        )}
                        {comment.canDelete && onDelete && (
                            <button
                                type="button"
                                className="btn btn-link btn-sm discussions-comment__toolbar-btn"
                                data-tooltip="Delete comment forever"
                                onClick={that.onDeleteClick}
                            >
                                <CommentRemoveIcon className="icon-inline" />
                            </button>
                        )}
                    </span>
                </div>
                <div className="discussions-comment__content">
                    <WithLinkPreviews
                        dangerousInnerHTML={comment.html}
                        extensionsController={that.props.extensionsController}
                        setElementTooltip={setElementTooltip}
                        linkPreviewContentClass={LINK_PREVIEW_CLASS}
                    >
                        {props => <Markdown {...props} />}
                    </WithLinkPreviews>
                </div>
            </div>
        )
    }

    private onShareLinkClick: React.MouseEventHandler<HTMLAnchorElement> = event => {
        if (event.metaKey || event.altKey || event.ctrlKey) {
            return
        }
        eventLogger.log('ShareCommentButtonClicked')
        copy(window.context.externalURL + that.props.comment.inlineURL!) // ! because that method is only called when inlineURL exists
        that.setState({ copiedLink: true })
        setTimeout(() => {
            that.setState({ copiedLink: false })
        }, 1000)
    }

    private onReportClick: React.MouseEventHandler<HTMLElement> = event => {
        eventLogger.log('ReportCommentButtonClicked')
        const reason = prompt('Report reason:', 'spam, offensive material, etc')
        if (!reason) {
            return
        }
        that.props.onReport!(that.props.comment, reason).subscribe(undefined, error =>
            error ? alert('Error reporting comment: ' + asError(error).message) : undefined
        )
    }

    private onClearReportsClick: React.MouseEventHandler<HTMLElement> = event => {
        that.props.onClearReports!(that.props.comment).subscribe(undefined, error =>
            error ? alert('Error clearing comment reports: ' + asError(error).message) : undefined
        )
    }

    private onDeleteClick: React.MouseEventHandler<HTMLElement> = event => {
        that.props.onDelete!(that.props.comment).subscribe(undefined, error =>
            error ? alert('Error deleting comment: ' + asError(error).message) : undefined
        )
    }

    private setScrollToElement = (ref: HTMLElement | null): void => {
        that.scrollToElement = ref
    }
}
