import * as H from 'history'
import * as React from 'react'
import { Link } from 'react-router-dom'
import { Observable } from 'rxjs'
import { ChatIcon } from '../../../shared/src/components/icons'
import * as GQL from '../../../shared/src/graphql/schema'
import { FilteredConnection, FilteredConnectionQueryArgs } from '../components/FilteredConnection'
import { Timestamp } from '../components/time/Timestamp'
import { fetchDiscussionThreads } from './backend'

interface DiscussionNodeProps {
    node: Pick<
        GQL.IDiscussionThread,
        'idWithoutKind' | 'title' | 'author' | 'inlineURL' | 'comments' | 'createdAt' | 'targets'
    >
    location: H.Location
    withRepo?: boolean
}

const DiscussionNode: React.FunctionComponent<DiscussionNodeProps> = ({ node, location, withRepo }) => {
    const currentURL = location.pathname + location.search + location.hash

    // TODO(slimsag:discussions): future: Improve rendering of discussions when there is no inline URL
    const inlineURL = node.inlineURL || ''

    // TODO(sqs): support linking to multiple targets
    const target =
        node.targets && node.targets.nodes && node.targets.nodes.length > 0 ? node.targets.nodes[0] : undefined

    return (
        <li className={'discussions-list__row' + (currentURL === inlineURL ? ' discussions-list__row--active' : '')}>
            <div className="d-flex align-items-center justify-content-between">
                <h3 className="discussions-list__row-title mb-0">
                    <Link to={inlineURL}>{node.title}</Link>
                </h3>
                <Link to={inlineURL} className="text-muted">
                    <ChatIcon className="icon-inline mr-1" />
                    {node.comments.totalCount}
                </Link>
            </div>
            <div className="text-muted">
                #{node.idWithoutKind} created <Timestamp date={node.createdAt} /> by{' '}
                <Link to={`/users/${node.author.username}`} data-tooltip={node.author.displayName}>
                    {node.author.username}
                </Link>{' '}
                {target && target.__typename == 'DiscussionThreadTargetRepo' && withRepo && (
                    <>
                        in <Link to={target.repository.name}>{target.repository.name}</Link>
                    </>
                )}
            </div>
        </li>
    )
}

class FilteredDiscussionsConnection extends FilteredConnection<
    DiscussionNodeProps['node'],
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
    compact: boolean
}

export class DiscussionsList extends React.PureComponent<Props> {
    public render(): JSX.Element | null {
        return (
            <FilteredDiscussionsConnection
                className={'discussions-list' + this.props.noFlex ? 'discussions-list--no-flex' : ''}
                autoFocus={this.props.autoFocus !== undefined ? this.props.autoFocus : true}
                compact={this.props.compact}
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
