import CaretDownIcon from '@sourcegraph/icons/lib/CaretDown'
import ComputerIcon from '@sourcegraph/icons/lib/Computer'
import CopyIcon from '@sourcegraph/icons/lib/Copy'
import GitHub from '@sourcegraph/icons/lib/GitHub'
import copy from 'copy-to-clipboard'
import * as H from 'history'
import * as React from 'react'
import { RepoBreadcrumb } from 'sourcegraph/components/Breadcrumb'
import { events } from 'sourcegraph/tracking/events'
import { parseHash, toEditorURL } from 'sourcegraph/util/url'

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
        return (
            <div className='repo-nav'>
                <span className='repo-nav__rev' onMouseDown={this.preventDefault} onClick={this.props.onClickRevision}>
                     {/* TODO(future): It's bad to assume master! We also do this below in this file, and in repo/backend.tsx  */}
                    <span className='repo-nav__rev-text'>{this.props.rev || 'master'}</span>
                    <CaretDownIcon />
                </span>
                <span className='repo-nav__path'>
                    <RepoBreadcrumb {...this.props} />
                </span>
                <a href='' className='repo-nav__action' onClick={this.onShareButtonClick}>
                    <CopyIcon />
                    {this.state.copiedLink ? 'Copied!' : 'Copy link'}
                </a>
                {
                    this.props.filePath && this.props.repoPath.split('/')[0] === 'github.com' &&
                        <a href={this.urlToGitHub()} target='_blank' className='repo-nav__action'>
                            <GitHub />
                            View on GitHub
                        </a>
                }
                {
                    this.props.repoPath &&
                        <a href={toEditorURL(this.props.repoPath, this.props.commitID, this.props.filePath, parseHash(this.props.location.hash))} target='_blank' className='repo-nav__action'>
                            <ComputerIcon />
                            <span>Open on desktop</span>
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

    /**
     * preventDefault is called when the mouseDown event is fired on the revision button. This
     * prevents mouseClick from firing when the RevSwitcher's blur event would also fire and
     * both toggle the same state in response to a single user mouse click.
     */
    private preventDefault = e => e.preventDefault()

    private urlToGitHub(): string {
        const hash = parseHash(this.props.location.hash)
        return `https://${this.props.repoPath}/blob/${this.props.rev || 'master'}/${this.props.filePath}${hash.line ? '#L' + hash.line : ''}`
    }
}
