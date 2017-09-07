import ListIcon from '@sourcegraph/icons/lib/List'
import RepoIcon from '@sourcegraph/icons/lib/Repo'
import ShareIcon from '@sourcegraph/icons/lib/Share'
import * as copy from 'copy-to-clipboard'
import * as H from 'history'
import * as React from 'react'
import * as GitHub from 'react-icons/lib/go/mark-github'
import { RepoBreadcrumb } from 'sourcegraph/components/Breadcrumb'
import { events } from 'sourcegraph/tracking/events'
import * as url from 'sourcegraph/util/url'
import * as URI from 'urijs'

interface RepoSubnavProps {
    repoPath: string
    rev?: string
    filePath?: string
    onClickNavigation: () => void
    location: H.Location
}

interface RepoSubnavState {
    copiedLink?: boolean
}

export class RepoNav extends React.Component<RepoSubnavProps, RepoSubnavState> {
    public state: RepoSubnavState = {}

    public render(): JSX.Element | null {
        const hash = url.parseHash(this.props.location.hash)
        return <div className='repo-nav'>
            <span className='explorer' onClick={this.props.onClickNavigation}>
                <ListIcon />
                Navigation
            </span>
            <span className='path'>
                <RepoIcon />
                <RepoBreadcrumb {...this.props} />
            </span>
            <span className='fill' />
            <span className='share' onClick={() => {
                events.ShareButtonClicked.log()

                const shareLink = URI.parse(window.location.href) // TODO(john): use this.props.location
                shareLink.query = (shareLink.query ? `${shareLink.query}&` : '') + 'utm_source=share'
                copy(URI.build(shareLink))
                this.setState({ copiedLink: true })

                setTimeout(() => {
                    this.setState({ copiedLink: undefined })
                }, 3000)
            }}>
                {this.state.copiedLink ? 'Copied link to clipboard!' : 'Share'}
                <ShareIcon />
            </span>
            {this.props.filePath && this.props.repoPath.split('/')[0] === 'github.com' &&
                <a href={url.toGitHubBlob({ uri: this.props.repoPath, rev: this.props.rev || 'master', path: this.props.filePath, line: hash.line })} className='view-external'>
                    View on GitHub
                <GitHub className='github-icon' /* TODO(john): use icon library */ />
                </a>}
        </div>
    }
}
