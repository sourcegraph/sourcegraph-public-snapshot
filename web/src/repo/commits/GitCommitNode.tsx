import CodeTagsIcon from '@sourcegraph/icons/lib/CodeTags'
import CopyIcon from '@sourcegraph/icons/lib/Copy'
import MoreIcon from '@sourcegraph/icons/lib/More'
import copy from 'copy-to-clipboard'
import * as React from 'react'
import { Link } from 'react-router-dom'
import { Timestamp } from '../../components/time/Timestamp'
import { Tooltip } from '../../components/tooltip/Tooltip'
import { eventLogger } from '../../tracking/eventLogger'
import { UserAvatar } from '../../user/UserAvatar'
import { externalCommitURL, toRepoURL } from '../../util/url'

const GitCommitNodeByline: React.SFC<{
    author: GQL.ISignature
    committer: GQL.ISignature | null
    className: string
}> = ({ author, committer, className }) => {
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
                <strong>{author.person.displayName}</strong> authored and{' '}
                <strong>{committer.person.displayName}</strong> committed <Timestamp date={committer.date} />
            </small>
        )
    }

    return (
        <small className={`git-commit-node-byline git-commit-node-byline--no-committer ${className}`}>
            <UserAvatar className="icon-inline mr-1" user={author.person} tooltip={author.person.displayName} />{' '}
            <strong>{author.person.displayName}</strong> committed <Timestamp date={author.date} />
        </small>
    )
}

export interface GitCommitNodeProps {
    node: GQL.IGitCommit

    /** The repository that contains this commit. */
    repoName: string

    /** An optional additional CSS class name to apply to this element. */
    className?: string
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
        return (
            <div key={this.props.node.id} className={`git-commit-node ${this.props.className}`}>
                <div className="git-commit-node__row git-commit-node__main">
                    <div className="git-commit-node__signature">
                        <div className="git-commit-node__message">
                            <Link
                                to={externalCommitURL(this.props.repoName, this.props.node.oid)}
                                className="git-commit-node__message-subject"
                                title={this.props.node.message}
                            >
                                {this.props.node.subject}
                            </Link>
                            {this.props.node.body && (
                                <button
                                    type="button"
                                    className="btn btn-secondary btn-sm git-commit-node__message-toggle"
                                    onClick={this.toggleShowCommitMessageBody}
                                >
                                    <MoreIcon className="icon-inline" />
                                </button>
                            )}
                        </div>
                        <GitCommitNodeByline
                            className="text-muted"
                            author={this.props.node.author}
                            committer={this.props.node.committer}
                        />
                    </div>
                    <div className="git-commit-node__actions">
                        <div className="btn-group btn-group-sm mr-2" role="group">
                            <Link
                                className="btn btn-outline-primary"
                                to={externalCommitURL(this.props.repoName, this.props.node.oid)}
                                data-tooltip="View this commit"
                            >
                                <strong>
                                    <code className="git-commit-node__oid">{this.props.node.abbreviatedOID}</code>
                                </strong>
                            </Link>
                            <button
                                type="button"
                                className="btn btn-outline-primary"
                                onClick={this.copyToClipboard}
                                data-tooltip={this.state.flashCopiedToClipboardMessage ? 'Copied!' : 'Copy full SHA'}
                            >
                                <CopyIcon className="icon-inline" />
                            </button>
                        </div>
                        <Link
                            className="btn btn-outline-primary btn-sm"
                            to={toRepoURL({ repoPath: this.props.repoName, rev: this.props.node.oid })}
                            data-tooltip="View files at this commit"
                        >
                            <CodeTagsIcon className="icon-inline" />
                        </Link>
                    </div>
                </div>
                {this.state.showCommitMessageBody && (
                    <div className="git-commit-node__row">
                        <pre className="git-commit-node__message-body">
                            <code>{this.props.node.body}</code>
                        </pre>
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
