import copy from 'copy-to-clipboard'
import formatDistance from 'date-fns/formatDistance'
import * as H from 'history'
import marked from 'marked'
import * as React from 'react'
import { Link } from 'react-router-dom'
import { Redirect } from 'react-router-dom'
import { UserAvatar } from '../settings/user/UserAvatar'

interface Props {
    comment: GQL.IComment
    location: H.Location
}

interface State {
    copiedLink: boolean
}

export class Comment extends React.Component<Props, State> {
    private scrollToElement: HTMLElement | null

    public state: State = {
        copiedLink: false,
    }

    public componentDidMount(): void {
        if (this.scrollToElement) {
            this.scrollToElement.scrollIntoView()
        }
    }

    public render(): JSX.Element | null {
        // If not logged in, redirect to sign in
        if (!window.context.user) {
            const newUrl = new URL(window.location.href)
            newUrl.pathname = '/sign-in'
            newUrl.searchParams.set('returnTo', window.location.href)
            return <Redirect to={newUrl.pathname + newUrl.search} />
        }

        const comment = this.props.comment
        const timeSince = formatDistance(comment.createdAt, new Date(), { addSuffix: true })
        const loc = this.props.location

        // Determine the (relative) URL to the comment.
        const shareUrl = new URL(loc.pathname + loc.search + loc.hash, window.location.href)
        shareUrl.searchParams.set('id', String(this.props.comment.id))
        const shareLinkHref = shareUrl.pathname + shareUrl.search + shareUrl.hash

        // Check if this comment is targeted.
        const isTargeted = new URLSearchParams(loc.search).get('id') === String(this.props.comment.id)

        return (
            <div className={`comment${isTargeted ? ' comment--targeted' : ''}`} ref={isTargeted ? this.setScrollToElement : undefined}>
                <div className='comment__top-area'>
                    <span className='comment__author' title={comment.author.username || undefined}>
                        <UserAvatar size={16} user={comment.author} className='comment__author-avatar' />
                        <span className='comment__author-name'>{comment.author.displayName ? comment.author.displayName : comment.author.username}</span>
                    </span>
                    <Link to={shareLinkHref} className='comment__share' title='Copy link to this comment' onClick={this.onShareLinkClick}>
                        {this.state.copiedLink ? 'Copied link to clipboard!' : timeSince}
                    </Link>
                </div>
                <div className='comment__content' dangerouslySetInnerHTML={{ __html: marked(comment.contents, { gfm: true, breaks: true, sanitize: true }) }}></div>
            </div>
        )
    }

    private onShareLinkClick: React.MouseEventHandler<HTMLElement> = event => {
        if (event.metaKey || event.altKey || event.ctrlKey) {
            return
        }
        const loc = this.props.location
        const shareLink = new URL(loc.pathname + loc.search + loc.hash, window.location.href)
        shareLink.searchParams.set('id', String(this.props.comment.id))
        copy(shareLink.href)
        this.setState({ copiedLink: true })

        setTimeout(() => {
            this.setState({ copiedLink: false })
        }, 1000)
    }

    private setScrollToElement = (ref: HTMLElement | null) => {
        this.scrollToElement = ref
    }
}
