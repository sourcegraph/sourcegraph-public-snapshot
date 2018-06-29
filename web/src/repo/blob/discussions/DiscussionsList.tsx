import ChatIcon from '@sourcegraph/icons/lib/Chat'
import * as H from 'history'
import * as React from 'react'
import { Link } from 'react-router-dom'
import { Observable } from 'rxjs'
import * as GQL from '../../../backend/graphqlschema'
import { FilteredConnection } from '../../../components/FilteredConnection'
import { Timestamp } from '../../../components/time/Timestamp'
import { UserAvatar } from '../../../user/UserAvatar'
import { formatHash, openFromJS } from '../../../util/url'
import { fetchDiscussionThreads } from './DiscussionsBackend'

interface DiscussionNodeProps {
    node: GQL.IDiscussionThread
    location: H.Location
}

const DiscussionNode: React.SFC<DiscussionNodeProps> = ({ node, location }) => {
    const currentURL = location.pathname + location.search + location.hash

    const hash = new URLSearchParams()
    hash.set('tab', 'discussions')
    hash.set('threadID', node.id)
    const hashString =
        node.target.__typename === 'DiscussionThreadTargetRepo' && node.target.selection !== null
            ? formatHash(
                  {
                      line: node.target.selection.startLine,
                      character: node.target.selection.startCharacter,
                      endLine: node.target.selection.endLine,
                      endCharacter: node.target.selection.endCharacter,
                  },
                  hash
              )
            : '#' + hash.toString()

    const discussionURL = location.pathname + location.search + hashString

    const openDiscussion = (e: any) => openFromJS(discussionURL, e)
    const preventDefault = (e: React.MouseEvent<HTMLAnchorElement>) => e.preventDefault()
    const stopPropagation = (e: React.MouseEvent<HTMLAnchorElement>) => e.stopPropagation()

    // TODO(slimsag:discussions): future: need to explain to user how to create
    // a thread in case they get here without selecting lines first.

    // TODO(slimsag:discussions): future: UserAvatar renders off-center when users do not have profile pictures.
    return (
        <li
            className={'discussions-list__row' + (currentURL === discussionURL ? ' discussions-list__row--active' : '')}
            onClick={openDiscussion}
        >
            <div className="discussions-list__row-top-line">
                <a href="#" onClick={preventDefault} className="discussions-list__row-title ">
                    {node.title}
                </a>
                <small className="discussions-list__row-id">#{node.id}</small>
                <span className="discussions-list__row-spacer" />
                <span
                    className="discussions-list__row-comments-count"
                    data-tooltip={node.comments.totalCount + ' comments'}
                >
                    <ChatIcon className="icon-inline" /> {node.comments.totalCount}
                </span>
            </div>
            <div className="discussions-list__row-bottom-line">
                <span>
                    Created <Timestamp date={node.createdAt} /> by{' '}
                    <Link
                        to={`/users/${node.author.username}`}
                        onClick={stopPropagation}
                        data-tooltip={node.author.displayName}
                    >
                        {node.author.username}
                    </Link>{' '}
                    <Link
                        to={`/users/${node.author.username}`}
                        onClick={stopPropagation}
                        data-tooltip={node.author.displayName}
                    >
                        <UserAvatar user={node.author} className="icon-inline icon-sm" />
                    </Link>
                </span>
            </div>
        </li>
    )
}

class FilteredDiscussionsConnection extends FilteredConnection<
    GQL.IDiscussionThread,
    Pick<DiscussionNodeProps, 'location'>
> {}

interface Props {
    repoID: GQL.ID
    rev: string | undefined
    filePath: string
    history: H.History
    location: H.Location
}

export class DiscussionsList extends React.PureComponent<Props> {
    public render(): JSX.Element | null {
        return (
            <FilteredDiscussionsConnection
                className="discussions-list"
                autoFocus={true}
                compact={true}
                noun="discussion"
                pluralNoun="discussions"
                queryConnection={this.fetchThreads}
                nodeComponent={DiscussionNode}
                nodeComponentProps={{ location: this.props.location } as Pick<DiscussionNodeProps, 'location'>}
                defaultFirst={100}
                shouldUpdateURLQuery={false}
                history={this.props.history}
                location={this.props.location}
            />
        )
    }

    private fetchThreads = (args: { query?: string }): Observable<GQL.IDiscussionThreadConnection> =>
        fetchDiscussionThreads()
}
