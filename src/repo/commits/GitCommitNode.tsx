import copy from 'copy-to-clipboard'
import ContentCopyIcon from 'mdi-react/ContentCopyIcon'
import DotsHorizontalIcon from 'mdi-react/DotsHorizontalIcon'
import FileDocumentIcon from 'mdi-react/FileDocumentIcon'
import * as React from 'react'
import { Link } from 'react-router-dom'
import * as GQL from '../../backend/graphqlschema'
import { Timestamp } from '../../components/time/Timestamp'
import { Tooltip } from '../../components/tooltip/Tooltip'
import { eventLogger } from '../../tracking/eventLogger'
import { UserAvatar } from '../../user/UserAvatar'
import { pluralize } from '../../util/strings'

const GitCommitNodeByline: React.SFC<{
    author: GQL.ISignature
    committer: GQL.ISignature | null
    className: string
    compact: boolean
}> = ({ author, committer, className, compact }) => {
    if (
        committer &&
        committer.person.email !== author.person.email &&
        ((!committer.person.name && !author.person.name) || committer.person.name !== author.person.name)
    ) {
        // The author and committer both exist and are different people.
        return (
            <small className={`git-commit-node-byline git-commit-node-byline--has-committer ${className}`}>
                <UserAvatar
                    className="icon-inline"
                    user={author.person}
                    tooltip={`${author.person.displayName} (author)`}
                />{' '}
                <UserAvatar
                    className="icon-inline mr-1"
                    user={committer.person}
                    tooltip={`${committer.person.displayName} (committer)`}
                />{' '}
                <strong>{author.person.displayName}</strong> {!compact && 'authored'} and{' '}
                <strong>{committer.person.displayName}</strong>{' '}
                {!compact && (
                    <>
                        committed <Timestamp date={committer.date} />
                    </>
                )}
            </small>
        )
    }

    return (
        <small className={`git-commit-node-byline git-commit-node-byline--no-committer ${className}`}>
            <UserAvatar className="icon-inline mr-1" user={author.person} tooltip={author.person.displayName} />{' '}
            <strong>{author.person.displayName}</strong>{' '}
            {!compact && (
                <>
                    committed <Timestamp date={author.date} />
                </>
            )}
        </small>
    )
}

export interface GitCommitNodeProps {
    node: GQL.IGitCommit

    /** The repository that contains this commit. */
    repoName: string

    /** An optional additional CSS class name to apply to this element. */
    className?: string

    /** Display in a single line (more compactly). */
    compact?: boolean

    /** Expand the commit message body. */
    expandCommitMessageBody?: boolean

    /** Show the full 40-character SHA and parents on their own row. */
    showSHAAndParentsRow?: boolean
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
                author={this.props.node.author}
                committer={this.props.node.committer}
                compact={Boolean(this.props.compact)}
            />
        )
        const messageElement = (
            <div className="git-commit-node__message">
                <Link
                    to={this.props.node.canonicalURL}
                    className="git-commit-node__message-subject"
                    title={this.props.node.message}
                >
                    {this.props.node.subject}
                </Link>
                {this.props.node.body &&
                    !this.props.expandCommitMessageBody && (
                        <button
                            type="button"
                            className="btn btn-secondary btn-sm git-commit-node__message-toggle"
                            onClick={this.toggleShowCommitMessageBody}
                        >
                            <DotsHorizontalIcon className="icon-inline" />
                        </button>
                    )}
                {this.props.compact && (
                    <small className="text-muted git-commit-node__message-timestamp">
                        <Timestamp
                            noAbout={true}
                            date={
                                this.props.node.committer ? this.props.node.committer.date : this.props.node.author.date
                            }
                        />
                    </small>
                )}
            </div>
        )
        const oidElement = <code className="git-commit-node__oid">{this.props.node.abbreviatedOID}</code>

        return (
            <div
                key={this.props.node.id}
                className={`git-commit-node ${this.props.compact ? 'git-commit-node--compact' : ''} ${this.props
                    .className || ''}`}
            >
                <div className="git-commit-node__row git-commit-node__main">
                    {!this.props.compact ? (
                        <>
                            <div className="git-commit-node__signature">
                                {messageElement}
                                {bylineElement}
                            </div>
                            <div className="git-commit-node__actions">
                                {!this.props.showSHAAndParentsRow && (
                                    <div className="btn-group btn-group-sm mr-2" role="group">
                                        <Link
                                            className="btn btn-secondary"
                                            to={this.props.node.canonicalURL}
                                            data-tooltip="View this commit"
                                        >
                                            <strong>{oidElement}</strong>
                                        </Link>
                                        <button
                                            type="button"
                                            className="btn btn-secondary"
                                            onClick={this.copyToClipboard}
                                            data-tooltip={
                                                this.state.flashCopiedToClipboardMessage ? 'Copied!' : 'Copy full SHA'
                                            }
                                        >
                                            <ContentCopyIcon className="icon-inline" />
                                        </button>
                                    </div>
                                )}
                                {this.props.node.tree && (
                                    <Link
                                        className="btn btn-secondary btn-sm"
                                        to={this.props.node.tree.canonicalURL}
                                        data-tooltip="View files at this commit"
                                    >
                                        <FileDocumentIcon className="icon-inline" />
                                    </Link>
                                )}
                            </div>
                        </>
                    ) : (
                        <>
                            {bylineElement}
                            {messageElement}
                            <Link to={this.props.node.canonicalURL}>{oidElement}</Link>
                        </>
                    )}
                </div>
                {(this.props.expandCommitMessageBody || this.state.showCommitMessageBody) && (
                    <div className="git-commit-node__row">
                        <pre className="git-commit-node__message-body">{this.props.node.body}</pre>
                    </div>
                )}
                {this.props.showSHAAndParentsRow && (
                    <div className="git-commit-node__row git-commit-node__sha-and-parents">
                        <code className="git-ref-tag-2 git-commit-node__sha-and-parents-sha">
                            {this.props.node.oid}{' '}
                            <button
                                type="button"
                                className="btn btn-icon git-commit-node__sha-and-parents-copy"
                                onClick={this.copyToClipboard}
                                data-tooltip={this.state.flashCopiedToClipboardMessage ? 'Copied!' : 'Copy full SHA'}
                            >
                                <ContentCopyIcon className="icon-inline" />
                            </button>
                        </code>
                        <div className="git-commit-node__sha-and-parents-parents">
                            {this.props.node.parents.length > 0 ? (
                                <>
                                    <span className="git-commit-node__sha-and-parents-label">
                                        {this.props.node.parents.length === 1
                                            ? 'Parent'
                                            : `${this.props.node.parents.length} ${pluralize(
                                                  'parent',
                                                  this.props.node.parents.length
                                              )}`}:
                                    </span>{' '}
                                    {this.props.node.parents.map((parent, i) => (
                                        <Link
                                            key={i}
                                            className="git-ref-tag-2 git-commit-node__sha-and-parents-parent"
                                            to={`/${this.props.repoName}/-/commit/${parent.oid}`}
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

    private toggleShowCommitMessageBody = () => {
        eventLogger.log('CommitBodyToggled')
        this.setState(prevState => ({ showCommitMessageBody: !prevState.showCommitMessageBody }))
    }

    private copyToClipboard = (): void => {
        eventLogger.log('CommitSHACopiedToClipboard')
        copy(this.props.node.oid)
        this.setState({ flashCopiedToClipboardMessage: true }, () => {
            Tooltip.forceUpdate()
            setTimeout(() => {
                this.setState({ flashCopiedToClipboardMessage: false }, () => Tooltip.forceUpdate())
            }, 1500)
        })
    }
}
