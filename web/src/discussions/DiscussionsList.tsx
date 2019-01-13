import * as H from 'history'
import * as React from 'react'
import { Link } from 'react-router-dom'
import { Observable } from 'rxjs'
import { ChatIcon } from '../../../shared/src/components/icons'
import * as GQL from '../../../shared/src/graphql/schema'
import { FilteredConnection, FilteredConnectionQueryArgs } from '../components/FilteredConnection'
import { Timestamp } from '../components/time/Timestamp'
import { openFromJS } from '../util/url'
import { fetchDiscussionThreads } from './backend'

interface DiscussionNodeProps {
    node: GQL.IDiscussionThread
    location: H.Location
    withRepo?: boolean
}

const DiscussionNode: React.FunctionComponent<DiscussionNodeProps> = ({ node, location, withRepo }) => {
    const currentURL = location.pathname + location.search + location.hash

    // TODO(slimsag:discussions): future: Improve rendering of discussions when there is no inline URL
    const inlineURL = node.inlineURL || ''

    const openDiscussion = (e: any) => openFromJS(inlineURL, e)
    const preventDefault = (e: React.MouseEvent<HTMLAnchorElement>) => e.preventDefault()
    const stopPropagation = (e: React.MouseEvent<HTMLAnchorElement>) => e.stopPropagation()

    console.log(withRepo)
    return (
        <li
            className={'discussions-list__row' + (currentURL === inlineURL ? ' discussions-list__row--active' : '')}
            onClick={openDiscussion}
        >
            <div className="discussions-list__row-top-line">
                <h3 className="discussions-list__row-title ">
                    <a href="#" onClick={preventDefault}>
                        {node.title}
                    </a>
                    <small className="discussions-list__row-id">#{node.id}</small>
                </h3>
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
                    />
                    <br />
                    {withRepo && (
                        <Link
                            to={node.target.repository.name}
                            onClick={stopPropagation}
                            data-tooltip={'View repository on Sourcegraph'}
                        >
                            {node.target.repository.name}
                        </Link>
                    )}
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
    withRepo?: boolean
}

export class DiscussionsList extends React.PureComponent<Props> {
    public render(): JSX.Element | null {
        return (
            <FilteredDiscussionsConnection
                className={'discussions-list' + this.props.noFlex ? 'discussions-list--no-flex' : ''}
                autoFocus={this.props.autoFocus !== undefined ? this.props.autoFocus : true}
                compact={false}
                noun={this.props.noun || 'discussion'}
                pluralNoun={this.props.pluralNoun || 'discussions'}
                queryConnection={this.fetchThreads}
                nodeComponent={DiscussionNode}
                nodeComponentProps={
                    { location: this.props.location, withRepo: this.props.withRepo } as Pick<
                        DiscussionNodeProps,
                        'location'
                    >
                }
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
