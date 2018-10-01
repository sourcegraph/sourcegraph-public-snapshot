import * as H from 'history'
import * as React from 'react'
import { Link } from 'react-router-dom'
import { Observable } from 'rxjs'
import * as GQL from '../backend/graphqlschema'
import { FilteredConnection, FilteredConnectionQueryArgs } from '../components/FilteredConnection'
import { Timestamp } from '../components/time/Timestamp'
import { UserAvatar } from '../user/UserAvatar'
import { ChatIcon } from '../util/icons'
import { openFromJS } from '../util/url'
import { fetchDiscussionThreads } from './backend'

interface DiscussionNodeProps {
    node: GQL.IDiscussionThread
    location: H.Location
}

const DiscussionNode: React.SFC<DiscussionNodeProps> = ({ node, location }) => {
    const currentURL = location.pathname + location.search + location.hash

    // TODO(slimsag:discussions): future: Improve rendering of discussions when there is no inline URL
    const inlineURL = node.inlineURL || ''

    const openDiscussion = (e: any) => openFromJS(inlineURL, e)
    const preventDefault = (e: React.MouseEvent<HTMLAnchorElement>) => e.preventDefault()
    const stopPropagation = (e: React.MouseEvent<HTMLAnchorElement>) => e.stopPropagation()

    return (
        <li
            className={'discussions-list__row' + (currentURL === inlineURL ? ' discussions-list__row--active' : '')}
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
    repoID: GQL.ID | undefined
    rev: string | undefined
    filePath: string | undefined
    history: H.History
    location: H.Location

    autoFocus?: boolean
    defaultFirst?: number
    hideSearch?: boolean
    noun?: string
    pluralNoun?: string
    noFlex?: boolean
}

export class DiscussionsList extends React.PureComponent<Props> {
    public render(): JSX.Element | null {
        return (
            <FilteredDiscussionsConnection
                className={'discussions-list' + this.props.noFlex ? 'discussions-list--no-flex' : ''}
                autoFocus={this.props.autoFocus !== undefined ? this.props.autoFocus : true}
                compact={true}
                noun={this.props.noun || 'discussion'}
                pluralNoun={this.props.pluralNoun || 'discussions'}
                queryConnection={this.fetchThreads}
                nodeComponent={DiscussionNode}
                nodeComponentProps={{ location: this.props.location } as Pick<DiscussionNodeProps, 'location'>}
                updateOnChange={`${this.props.repoID}:${this.props.rev}:${this.props.filePath}`}
                defaultFirst={this.props.defaultFirst || 100}
                hideSearch={this.props.hideSearch}
                shouldUpdateURLQuery={false}
                history={this.props.history}
                location={this.props.location}
            />
        )
    }

    private fetchThreads = (args: FilteredConnectionQueryArgs): Observable<GQL.IDiscussionThreadConnection> =>
        fetchDiscussionThreads({
            ...args,
            targetRepositoryID: this.props.repoID,
            targetRepositoryPath: this.props.filePath,
        })
}
