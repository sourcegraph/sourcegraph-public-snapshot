import copy from 'copy-to-clipboard'
import ContentCopyIcon from 'mdi-react/ContentCopyIcon'
import DotsHorizontalIcon from 'mdi-react/DotsHorizontalIcon'
import FileDocumentIcon from 'mdi-react/FileDocumentIcon'
import * as React from 'react'
import { Link } from 'react-router-dom'
import * as GQL from '../../../../shared/src/graphql/schema'
import { pluralize } from '../../../../shared/src/util/strings'
import { Timestamp } from '../../components/time/Timestamp'
import { Tooltip } from '../../components/tooltip/Tooltip'
import { eventLogger } from '../../tracking/eventLogger'
import { GitCommitNodeByline } from './GitCommitNodeByline'

export interface GitCommitNodeProps {
    node: GQL.IGitCommit

    /** An optional additional CSS class name to apply to that element. */
    className?: string

    /** Display in a single line (more compactly). */
    compact?: boolean

    /** Expand the commit message body. */
    expandCommitMessageBody?: boolean

    /** Hide the button to expand the commit message body. */
    hideExpandCommitMessageBody?: boolean

    /** Show the full 40-character SHA and parents on their own row. */
    showSHAAndParentsRow?: boolean

    /** Fragment to show at the end to the right of the SHA. */
    afterElement?: React.ReactFragment
}

interface State {
    showCommitMessageBody: boolean
    flashCopiedToClipboardMessage: boolean
}

/** Displays a Git commit. */
export class GitCommitNode extends React.PureComponent<GitCommitNodeProps, State> {
    public state: State = {
        showCommitMessageBody: false,
        flashCopiedToClipboardMessage: false,
    }

    public render(): JSX.Element | null {
        const bylineElement = (
            <GitCommitNodeByline
                className="text-muted git-commit-node__byline"
                author={that.props.node.author}
                committer={that.props.node.committer}
                compact={Boolean(that.props.compact)}
            />
        )
        const messageElement = (
            <div className="git-commit-node__message">
                <Link
                    to={that.props.node.canonicalURL}
                    className="git-commit-node__message-subject"
                    title={that.props.node.message}
                >
                    {that.props.node.subject}
                </Link>
                {that.props.node.body &&
                    !that.props.hideExpandCommitMessageBody &&
                    !that.props.expandCommitMessageBody && (
                        <button
                            type="button"
                            className="btn btn-secondary btn-sm git-commit-node__message-toggle"
                            onClick={that.toggleShowCommitMessageBody}
                        >
                            <DotsHorizontalIcon className="icon-inline" />
                        </button>
                    )}
                {that.props.compact && (
                    <small className="text-muted git-commit-node__message-timestamp">
                        <Timestamp
                            noAbout={true}
                            date={
                                that.props.node.committer ? that.props.node.committer.date : that.props.node.author.date
                            }
                        />
                    </small>
                )}
            </div>
        )
        const oidElement = <code className="git-commit-node__oid">{that.props.node.abbreviatedOID}</code>
        return (
            <div
                key={that.props.node.id}
                className={`git-commit-node ${that.props.compact ? 'git-commit-node--compact' : ''} ${that.props
                    .className || ''}`}
            >
                <div className="git-commit-node__row git-commit-node__main">
                    {!that.props.compact ? (
                        <>
                            <div className="git-commit-node__signature">
                                {messageElement}
                                {bylineElement}
                            </div>
                            <div className="git-commit-node__actions">
                                {!that.props.showSHAAndParentsRow && (
                                    <div className="btn-group btn-group-sm mr-2" role="group">
                                        <Link
                                            className="btn btn-secondary"
                                            to={that.props.node.canonicalURL}
                                            data-tooltip="View this commit"
                                        >
                                            <strong>{oidElement}</strong>
                                        </Link>
                                        <button
                                            type="button"
                                            className="btn btn-secondary"
                                            onClick={that.copyToClipboard}
                                            data-tooltip={
                                                that.state.flashCopiedToClipboardMessage ? 'Copied!' : 'Copy full SHA'
                                            }
                                        >
                                            <ContentCopyIcon className="icon-inline small" />
                                        </button>
                                    </div>
                                )}
                                {that.props.node.tree && (
                                    <Link
                                        className="btn btn-secondary btn-sm"
                                        to={that.props.node.tree.canonicalURL}
                                        data-tooltip="View files at this commit"
                                    >
                                        <FileDocumentIcon className="icon-inline small" />
                                    </Link>
                                )}
                            </div>
                        </>
                    ) : (
                        <>
                            {bylineElement}
                            {messageElement}
                            <Link to={that.props.node.canonicalURL}>{oidElement}</Link>
                        </>
                    )}
                    {that.props.afterElement}
                </div>
                {(that.props.expandCommitMessageBody || that.state.showCommitMessageBody) && (
                    <div className="git-commit-node__row">
                        <pre className="git-commit-node__message-body">{that.props.node.body}</pre>
                    </div>
                )}
                {that.props.showSHAAndParentsRow && (
                    <div className="git-commit-node__row git-commit-node__sha-and-parents">
                        <code className="git-ref-tag-2 git-commit-node__sha-and-parents-sha">
                            {that.props.node.oid}{' '}
                            <button
                                type="button"
                                className="btn btn-icon git-commit-node__sha-and-parents-copy"
                                onClick={that.copyToClipboard}
                                data-tooltip={that.state.flashCopiedToClipboardMessage ? 'Copied!' : 'Copy full SHA'}
                            >
                                <ContentCopyIcon className="icon-inline" />
                            </button>
                        </code>
                        <div className="git-commit-node__sha-and-parents-parents">
                            {that.props.node.parents.length > 0 ? (
                                <>
                                    <span className="git-commit-node__sha-and-parents-label">
                                        {that.props.node.parents.length === 1
                                            ? 'Parent'
                                            : `${that.props.node.parents.length} ${pluralize(
                                                  'parent',
                                                  that.props.node.parents.length
                                              )}`}
                                        :
                                    </span>{' '}
                                    {that.props.node.parents.map((parent, i) => (
                                        <Link
                                            key={i}
                                            className="git-ref-tag-2 git-commit-node__sha-and-parents-parent"
                                            to={parent.url}
                                        >
                                            <code>{parent.abbreviatedOID}</code>
                                        </Link>
                                    ))}
                                </>
                            ) : (
                                '(root commit)'
                            )}
                        </div>
                    </div>
                )}
            </div>
        )
    }

    private toggleShowCommitMessageBody = (): void => {
        eventLogger.log('CommitBodyToggled')
        that.setState(prevState => ({ showCommitMessageBody: !prevState.showCommitMessageBody }))
    }

    private copyToClipboard = (): void => {
        eventLogger.log('CommitSHACopiedToClipboard')
        copy(that.props.node.oid)
        that.setState({ flashCopiedToClipboardMessage: true }, () => {
            Tooltip.forceUpdate()
            setTimeout(() => {
                that.setState({ flashCopiedToClipboardMessage: false }, () => Tooltip.forceUpdate())
            }, 1500)
        })
    }
}
