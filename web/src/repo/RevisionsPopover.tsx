import * as H from 'history'
import * as React from 'react'
import { Link } from 'react-router-dom'
import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'
import { CircleChevronLeftIcon } from '../../../shared/src/components/icons'
import { TabsWithLocalStorageViewStatePersistence } from '../../../shared/src/components/Tabs'
import { gql, dataOrThrowErrors } from '../../../shared/src/graphql/graphql'
import * as GQL from '../../../shared/src/graphql/schema'
import { memoizeObservable } from '../../../shared/src/util/memoizeObservable'
import { queryGraphQL } from '../backend/graphql'
import { FilteredConnection, FilteredConnectionQueryArgs } from '../components/FilteredConnection'
import { eventLogger } from '../tracking/eventLogger'
import { replaceRevisionInURL } from '../util/url'
import { GitRefNode, queryGitRefs } from './GitRef'

const fetchRepositoryCommits = memoizeObservable(
    (args: { repo: GQL.ID; rev?: string; first?: number; query?: string }): Observable<GQL.IGitCommitConnection> =>
        queryGraphQL(
            gql`
                query RepositoryGitCommit($repo: ID!, $first: Int, $rev: String!, $query: String) {
                    node(id: $repo) {
                        __typename
                        ... on Repository {
                            commit(rev: $rev) {
                                ancestors(first: $first, query: $query) {
                                    nodes {
                                        id
                                        oid
                                        abbreviatedOID
                                        author {
                                            person {
                                                name
                                                avatarURL
                                            }
                                            date
                                        }
                                        subject
                                    }
                                    pageInfo {
                                        hasNextPage
                                    }
                                }
                            }
                        }
                    }
                }
            `,
            args
        ).pipe(
            map(dataOrThrowErrors),
            map(({ node }) => {
                if (!node) {
                    throw new Error(`Repository ${args.repo} not found`)
                }
                if (node.__typename !== 'Repository') {
                    throw new Error(`Node is a ${node.__typename}, not a Repository`)
                }
                if (!node.commit?.ancestors) {
                    throw new Error(`Cannot load ancestors for repository ${args.repo}`)
                }
                return node.commit.ancestors
            })
        ),
    x => JSON.stringify(x)
)

interface GitRefPopoverNodeProps {
    node: GQL.IGitRef

    defaultBranch: string
    currentRev: string | undefined

    location: H.Location
}

const GitRefPopoverNode: React.FunctionComponent<GitRefPopoverNodeProps> = ({
    node,
    defaultBranch,
    currentRev,
    location,
}) => {
    let isCurrent: boolean
    if (currentRev) {
        isCurrent = node.name === currentRev || node.abbrevName === currentRev
    } else {
        isCurrent = node.name === `refs/heads/${defaultBranch}`
    }
    return (
        <GitRefNode
            node={node}
            url={replaceRevisionInURL(location.pathname + location.search + location.hash, node.abbrevName)}
            ancestorIsLink={false}
        >
            {isCurrent && (
                <CircleChevronLeftIcon
                    className="icon-inline connection-popover__node-link-icon"
                    data-tooltip="Current"
                />
            )}
        </GitRefNode>
    )
}

interface GitCommitNodeProps {
    node: GQL.IGitCommit

    currentCommitID: string | undefined

    location: H.Location
}

const GitCommitNode: React.FunctionComponent<GitCommitNodeProps> = ({ node, currentCommitID, location }) => {
    const isCurrent = currentCommitID === node.oid
    return (
        <li key={node.oid} className="connection-popover__node revisions-popover-git-commit-node">
            <Link
                to={replaceRevisionInURL(location.pathname + location.search + location.hash, node.oid)}
                className={`connection-popover__node-link ${
                    isCurrent ? 'connection-popover__node-link--active' : ''
                } revisions-popover-git-commit-node__link`}
            >
                <code className="revisions-popover-git-commit-node__oid" title={node.oid}>
                    {node.abbreviatedOID}
                </code>
                <span className="revisions-popover-git-commit-node__message">{(node.subject || '').slice(0, 200)}</span>
                {isCurrent && (
                    <CircleChevronLeftIcon
                        className="icon-inline connection-popover__node-link-icon"
                        data-tooltip="Current commit"
                    />
                )}
            </Link>
        </li>
    )
}

interface Props {
    repo: GQL.ID
    repoName: string
    defaultBranch: string

    /** The current revision, or undefined for the default branch. */
    currentRev: string | undefined

    currentCommitID?: string

    history: H.History
    location: H.Location
}

type RevisionsPopoverTabID = 'branches' | 'tags' | 'commits'

interface RevisionsPopoverTab {
    id: RevisionsPopoverTabID
    label: string
    noun: string
    pluralNoun: string
    type?: GQL.GitRefType
}

class FilteredGitRefConnection extends FilteredConnection<
    GQL.IGitRef,
    Pick<GitRefPopoverNodeProps, 'defaultBranch' | 'currentRev' | 'location'>
> {}

class FilteredGitCommitConnection extends FilteredConnection<
    GQL.IGitCommit,
    Pick<GitCommitNodeProps, 'currentCommitID' | 'location'>
> {}

/**
 * A popover that displays a searchable list of revisions (grouped by type) for
 * the current repository.
 */
export class RevisionsPopover extends React.PureComponent<Props> {
    private static LAST_TAB_STORAGE_KEY = 'RevisionsPopover.lastTab'

    private static TABS: RevisionsPopoverTab[] = [
        { id: 'branches', label: 'Branches', noun: 'branch', pluralNoun: 'branches', type: GQL.GitRefType.GIT_BRANCH },
        { id: 'tags', label: 'Tags', noun: 'tag', pluralNoun: 'tags', type: GQL.GitRefType.GIT_TAG },
        { id: 'commits', label: 'Commits', noun: 'commit', pluralNoun: 'commits' },
    ]

    public componentDidMount(): void {
        eventLogger.logViewEvent('RevisionsPopover')
    }

    public render(): JSX.Element | null {
        return (
            <div className="revisions-popover connection-popover">
                <TabsWithLocalStorageViewStatePersistence
                    tabs={RevisionsPopover.TABS}
                    storageKey={RevisionsPopover.LAST_TAB_STORAGE_KEY}
                    className="revisions-popover__tabs"
                >
                    {RevisionsPopover.TABS.map(tab =>
                        tab.type ? (
                            <FilteredGitRefConnection
                                key={tab.id}
                                className="connection-popover__content"
                                showMoreClassName="connection-popover__show-more"
                                compact={true}
                                noun={tab.noun}
                                pluralNoun={tab.pluralNoun}
                                queryConnection={
                                    tab.type === GQL.GitRefType.GIT_BRANCH ? this.queryGitBranches : this.queryGitTags
                                }
                                nodeComponent={GitRefPopoverNode}
                                nodeComponentProps={
                                    {
                                        defaultBranch: this.props.defaultBranch,
                                        currentRev: this.props.currentRev,
                                        location: this.props.location,
                                    } as Pick<GitRefPopoverNodeProps, 'defaultBranch' | 'currentRev' | 'location'>
                                }
                                defaultFirst={50}
                                autoFocus={true}
                                noSummaryIfAllNodesVisible={true}
                                useURLQuery={false}
                                history={this.props.history}
                                location={this.props.location}
                            />
                        ) : (
                            <FilteredGitCommitConnection
                                key={tab.id}
                                className="connection-popover__content"
                                compact={true}
                                noun={tab.noun}
                                pluralNoun={tab.pluralNoun}
                                queryConnection={this.queryRepositoryCommits}
                                nodeComponent={GitCommitNode}
                                nodeComponentProps={
                                    {
                                        currentCommitID: this.props.currentCommitID,
                                        location: this.props.location,
                                    } as Pick<GitCommitNodeProps, 'currentCommitID' | 'location'>
                                }
                                defaultFirst={15}
                                autoFocus={true}
                                history={this.props.history}
                                location={this.props.location}
                                noSummaryIfAllNodesVisible={true}
                                useURLQuery={false}
                            />
                        )
                    )}
                </TabsWithLocalStorageViewStatePersistence>
            </div>
        )
    }

    private queryGitBranches = (args: FilteredConnectionQueryArgs): Observable<GQL.IGitRefConnection> =>
        queryGitRefs({ ...args, repo: this.props.repo, type: GQL.GitRefType.GIT_BRANCH, withBehindAhead: false })

    private queryGitTags = (args: FilteredConnectionQueryArgs): Observable<GQL.IGitRefConnection> =>
        queryGitRefs({ ...args, repo: this.props.repo, type: GQL.GitRefType.GIT_TAG, withBehindAhead: false })

    private queryRepositoryCommits = (args: FilteredConnectionQueryArgs): Observable<GQL.IGitCommitConnection> =>
        fetchRepositoryCommits({
            ...args,
            repo: this.props.repo,
            rev: this.props.currentRev || this.props.defaultBranch,
        })
}
