import ListIcon from '@sourcegraph/icons/lib/List'
import RepoIcon from '@sourcegraph/icons/lib/Repo'
import ShareIcon from '@sourcegraph/icons/lib/Share'
import copy from 'copy-to-clipboard'
import * as H from 'history'
import * as React from 'react'
import GitHub from 'react-icons/lib/go/mark-github'
import { RepoBreadcrumb } from 'sourcegraph/components/Breadcrumb'
import { events } from 'sourcegraph/tracking/events'
import { parseHash } from 'sourcegraph/util/url'

interface RepoSubnavProps {
    repoPath: string
    rev?: string
    filePath?: string
    onClickNavigation?: () => void
    location: H.Location
}

interface RepoSubnavState {
    copiedLink?: boolean
}

export class RepoNav extends React.Component<RepoSubnavProps, RepoSubnavState> {
    public state: RepoSubnavState = {}

    public render(): JSX.Element | null {
        return (
            <div className='repo-nav'>
                <span className='explorer' onClick={this.props.onClickNavigation}>
                    <ListIcon />
                    Navigation
                </span>
                <span className='path'>
                    <RepoIcon />
                    <RepoBreadcrumb {...this.props} />
                </span>
                <span className='fill' />
                <span className='share' onClick={this.onShareButtonClick}>
                    {this.state.copiedLink ? 'Copied link to clipboard!' : 'Share'}
                    <ShareIcon />
                </span>
                {this.props.filePath && this.props.repoPath.split('/')[0] === 'github.com' &&
                    <a href={this.urlToGitHub()} className='view-external'>
                        View on GitHub
                    <GitHub className='github-icon' /* TODO(john): use icon library */ />
                    </a>}
            </div>
        )
    }

    private onShareButtonClick: React.MouseEventHandler<HTMLButtonElement> = () => {
        events.ShareButtonClicked.log()
        const loc = this.props.location
        const shareLink = new URL(loc.pathname + loc.search + loc.hash, window.location.href)
        shareLink.searchParams.set('utm_source', 'share')
        copy(shareLink.href)
        this.setState({ copiedLink: true })

        setTimeout(() => {
            this.setState({ copiedLink: undefined })
        }, 3000)
    }

    private urlToGitHub(): string {
        const hash = parseHash(this.props.location.hash)
        return `https://${this.props.repoPath}/blob/${this.props.rev || 'master'}/${this.props.filePath}${hash.line ? '#L' + hash.line : ''}`
    }
}
