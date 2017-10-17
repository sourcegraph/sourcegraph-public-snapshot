import ComputerIcon from '@sourcegraph/icons/lib/Computer'
import CopyIcon from '@sourcegraph/icons/lib/Copy'
import GitHubIcon from '@sourcegraph/icons/lib/GitHub'
import copy from 'copy-to-clipboard'
import * as H from 'history'
import * as React from 'react'
import { RepoBreadcrumb } from '../components/Breadcrumb'
import { events } from '../tracking/events'
import { parseHash, toEditorURL } from '../util/url'
import { RevSwitcher } from './RevSwitcher'

interface RepoSubnavProps {
    repoPath: string
    rev?: string
    commitID?: string
    filePath?: string
    onClickRevision?: () => void
    hideCopyLink?: boolean
    showOpenOnDesktop?: boolean
    customEditorURL?: string
    revSwitcherDisabled?: boolean
    breadcrumbDisabled?: boolean
    location: H.Location
    history: H.History
}

interface RepoSubnavState {
    copiedLink: boolean
}

export class RepoNav extends React.Component<RepoSubnavProps, RepoSubnavState> {
    public state: RepoSubnavState = {
        copiedLink: false,
    }

    public render(): JSX.Element | null {
        const editorUrl = this.props.customEditorURL || toEditorURL(this.props.repoPath, this.props.commitID, this.props.filePath, parseHash(this.props.location.hash))
        return (
            <div className='repo-nav'>
                {/* TODO Don't assume master! */}
                <RevSwitcher history={this.props.history} rev={this.props.rev || 'master'} repoPath={this.props.repoPath} disabled={this.props.revSwitcherDisabled} />
                <span className='repo-nav__path'>
                    <RepoBreadcrumb {...this.props} disabled={this.props.breadcrumbDisabled} />
                </span>
                {!this.props.hideCopyLink && <a href='' className='repo-nav__action' onClick={this.onShareButtonClick} title='Copy link'>
                    <CopyIcon className='icon-inline'/>
                    <span className='repo-nav__action-text'>{this.state.copiedLink ? 'Copied!' : 'Copy link'}</span>
                </a>}
                {
                    this.props.filePath && this.props.repoPath.split('/')[0] === 'github.com' &&
                        <a href={this.urlToGitHub()} target='_blank' className='repo-nav__action' title='View on GitHub' onClick={this.onViewOnGitHubButtonClicked}>
                            <GitHubIcon className='icon-inline' />
                            <span className='repo-nav__action-text'>View on GitHub</span>
                        </a>
                }
                {
                    /* TODO(john): remove showOpenOnDesktop alltogether when we're ready to show
                       desktop to users everywhere (see https://github.com/sourcegraph/sourcegraph/issues/7297) */
                    this.props.repoPath && this.props.showOpenOnDesktop &&
                        <a href={editorUrl} target='sourcegraphapp' className='repo-nav__action' title='Open on desktop' onClick={this.onOpenOnDesktopClicked}>
                            <ComputerIcon className='icon-inline'/>
                            <span className='repo-nav__action-text'>Open on desktop</span>
                        </a>
                }
            </div>
        )
    }

    private onShareButtonClick: React.MouseEventHandler<HTMLElement> = event => {
        event.preventDefault()
        events.ShareButtonClicked.log()
        const loc = this.props.location
        const shareLink = new URL(loc.pathname + loc.search + loc.hash, window.location.href)
        shareLink.searchParams.set('utm_source', 'share')
        copy(shareLink.href)
        this.setState({ copiedLink: true })

        setTimeout(() => {
            this.setState({ copiedLink: false })
        }, 1000)
    }

    private onViewOnGitHubButtonClicked: React.MouseEventHandler<HTMLAnchorElement> = () => {
        events.OpenInCodeHostClicked.log()
    }

    private onOpenOnDesktopClicked: React.MouseEventHandler<HTMLAnchorElement> = () => {
        events.OpenInNativeAppClicked.log()
    }

    private urlToGitHub(): string {
        const hash = parseHash(this.props.location.hash)
        return `https://${this.props.repoPath}/blob/${this.props.rev || 'master'}/${this.props.filePath}${hash.line ? '#L' + hash.line : ''}`
    }
}
