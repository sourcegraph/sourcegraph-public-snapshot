import ComputerIcon from '@sourcegraph/icons/lib/Computer'
import CopyIcon from '@sourcegraph/icons/lib/Copy'
import GitHubIcon from '@sourcegraph/icons/lib/GitHub'
import PhabricatorIcon from '@sourcegraph/icons/lib/Phabricator'
import copy from 'copy-to-clipboard'
import * as H from 'history'
import * as React from 'react'
import { Subscription } from 'rxjs/Subscription'
import { currentUser } from '../auth'
import { RepoBreadcrumb } from '../components/Breadcrumb'
import { eventLogger } from '../tracking/eventLogger'
import { parseHash, toEditorURL } from '../util/url'
import { RevSwitcher } from './RevSwitcher'

interface RepoSubnavProps {
    repoPath: string
    rev: string
    commitID?: string
    filePath?: string
    onClickRevision?: () => void
    hideCopyLink?: boolean
    customEditorURL?: string
    revSwitcherDisabled?: boolean
    breadcrumbDisabled?: boolean
    phabricatorCallsign?: string
    isDirectory: boolean
    /**
     * overrides the line number that 'View on GitHub' should link to. By
     * default, it is parsed from the current URL hash.
     */
    line?: number
    location: H.Location
    history: H.History
}

interface RepoSubnavState {
    copiedLink: boolean
    editorBeta: boolean
}

export class RepoNav extends React.PureComponent<RepoSubnavProps, RepoSubnavState> {
    private subscriptions = new Subscription()
    public state: RepoSubnavState = {
        copiedLink: false,
        editorBeta: false,
    }

    public componentDidMount(): void {
        this.subscriptions.add(
            currentUser.subscribe(user => {
                this.setState({ editorBeta: !!user && user.tags && user.tags.some(tag => tag.name === 'editor-beta') })
            })
        )
    }

    public render(): JSX.Element | null {
        const editorUrl =
            this.props.customEditorURL ||
            toEditorURL(
                this.props.repoPath,
                this.props.commitID,
                this.props.filePath,
                parseHash(this.props.location.hash)
            )
        return (
            <div className="repo-nav">
                <RevSwitcher
                    history={this.props.history}
                    rev={this.props.rev}
                    repoPath={this.props.repoPath}
                    disabled={this.props.revSwitcherDisabled}
                />
                <span className="repo-nav__path">
                    <RepoBreadcrumb {...this.props} disableLinks={this.props.breadcrumbDisabled} />
                </span>
                {!this.props.hideCopyLink && (
                    <a href="" className="repo-nav__action" onClick={this.onShareButtonClick} title="Copy link">
                        <CopyIcon className="icon-inline" />
                        <span className="repo-nav__action-text">{this.state.copiedLink ? 'Copied!' : 'Copy link'}</span>
                    </a>
                )}
                {this.props.repoPath.split('/')[0] === 'github.com' && (
                    <a
                        href={this.urlToGitHub()}
                        target="_blank"
                        className="repo-nav__action"
                        title="View on GitHub"
                        onClick={this.onViewOnCodeHostButtonClicked}
                    >
                        <GitHubIcon className="icon-inline" />
                        <span className="repo-nav__action-text">View on GitHub</span>
                    </a>
                )}
                {this.props.filePath &&
                    this.props.phabricatorCallsign && (
                        <a
                            href={this.urlToPhabricator()}
                            target="_blank"
                            className="repo-nav__action"
                            title="View on Phabricator"
                            onClick={this.onViewOnCodeHostButtonClicked}
                        >
                            <PhabricatorIcon className="icon-inline" />
                            <span className="repo-nav__action-text">View on Phabricator</span>
                        </a>
                    )}
                {this.props.repoPath &&
                    this.state.editorBeta && (
                        <a
                            href={editorUrl}
                            target="sourcegraphapp"
                            className="repo-nav__action"
                            title="Open in Sourcegraph Editor"
                            onClick={this.onOpenOnDesktopClicked}
                        >
                            <ComputerIcon className="icon-inline" />
                            <span className="repo-nav__action-text">Open in Sourcegraph Editor</span>
                        </a>
                    )}
            </div>
        )
    }

    private onShareButtonClick: React.MouseEventHandler<HTMLElement> = event => {
        event.preventDefault()
        eventLogger.log('ShareButtonClicked')
        const loc = this.props.location
        const shareLink = new URL(loc.pathname + loc.search + loc.hash, window.location.href)
        shareLink.searchParams.set('utm_source', 'share')
        copy(shareLink.href)
        this.setState({ copiedLink: true })

        setTimeout(() => {
            this.setState({ copiedLink: false })
        }, 1000)
    }

    private onViewOnCodeHostButtonClicked: React.MouseEventHandler<HTMLAnchorElement> = () => {
        eventLogger.log('OpenInCodeHostClicked')
    }

    private onOpenOnDesktopClicked: React.MouseEventHandler<HTMLAnchorElement> = () => {
        eventLogger.log('OpenInNativeAppClicked')
    }

    private urlToGitHub(): string {
        const line = this.props.line || parseHash(this.props.location.hash).line || undefined
        if (this.props.filePath) {
            if (this.props.isDirectory) {
                return `https://${this.props.repoPath}/tree/${this.props.rev}/${this.props.filePath}`
            }
            const url = new URL(`https://${this.props.repoPath}/blob/${this.props.rev}/${this.props.filePath}`)
            if (line) {
                url.hash = '#L' + line
            }
            return url.href
        }
        return `https://${this.props.repoPath}/tree/${this.props.rev}/`
    }

    private urlToPhabricator(): string {
        if (!window.context.phabricatorURL) {
            throw new Error('cannot locate Phabricator instance, make sure your admin has set PHABRICATOR_URL')
        }
        return `${window.context.phabricatorURL}/source/${this.props.phabricatorCallsign}/browse/${this.props.filePath}`
    }
}
