import CaretDownIcon from '@sourcegraph/icons/lib/CaretDown'
import ComputerIcon from '@sourcegraph/icons/lib/Computer'
import CopyIcon from '@sourcegraph/icons/lib/Copy'
import GitHub from '@sourcegraph/icons/lib/GitHub'
import copy from 'copy-to-clipboard'
import * as H from 'history'
import * as React from 'react'
import { RepoBreadcrumb } from '../components/Breadcrumb'
import { events } from '../tracking/events'
import { parseHash, toEditorURL } from '../util/url'

interface RepoSubnavProps {
    repoPath: string
    rev?: string
    commitID?: string
    filePath?: string
    onClickRevision?: () => void
    location: H.Location
}

interface RepoSubnavState {
    copiedLink: boolean
}

export class RepoNav extends React.Component<RepoSubnavProps, RepoSubnavState> {
    public state: RepoSubnavState = {
        copiedLink: false
    }

    public render(): JSX.Element | null {
        const editorUrl = toEditorURL(this.props.repoPath, this.props.commitID, this.props.filePath, parseHash(this.props.location.hash))
        return (
            <div className='repo-nav'>
                <span className='repo-nav__rev' onClick={this.props.onClickRevision}>
                     {/* TODO(future): It's bad to assume master! We also do this below in this file, and in repo/backend.tsx  */}
                    <span className='repo-nav__rev-text'>{this.props.rev || 'master'}</span>
                    <CaretDownIcon />
                </span>
                <span className='repo-nav__path'>
                    <RepoBreadcrumb {...this.props} />
                </span>
                <a href='' className='repo-nav__action' onClick={this.onShareButtonClick} title='Copy link'>
                    <CopyIcon />
                    <span className='repo-nav__action-text'>{this.state.copiedLink ? 'Copied!' : 'Copy link'}</span>
                </a>
                {
                    this.props.filePath && this.props.repoPath.split('/')[0] === 'github.com' &&
                        <a href={this.urlToGitHub()} target='_blank' className='repo-nav__action' title='View on GitHub'>
                            <GitHub />
                            <span className='repo-nav__action-text'>View on GitHub</span>
                        </a>
                }
                {
                    this.props.repoPath &&
                        <a href={editorUrl} target='_blank' className='repo-nav__action' title='Open on desktop'>
                            <ComputerIcon />
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

    private urlToGitHub(): string {
        const hash = parseHash(this.props.location.hash)
        return `https://${this.props.repoPath}/blob/${this.props.rev || 'master'}/${this.props.filePath}${hash.line ? '#L' + hash.line : ''}`
    }
}
