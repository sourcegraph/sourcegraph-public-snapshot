import { escapeRegExp, isEqual } from 'lodash'
import * as React from 'react'
import { Link, RouteComponentProps } from 'react-router-dom'
import { Observable, Subject } from 'rxjs'
import { map } from 'rxjs/operators'
import { gql } from '../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../shared/src/graphql/schema'
import { createAggregateError } from '../../../../shared/src/util/errors'
import { memoizeObservable } from '../../../../shared/src/util/memoizeObservable'
import { numberWithCommas, pluralize } from '../../../../shared/src/util/strings'
import { buildSearchURLQuery } from '../../../../shared/src/util/url'
import { queryGraphQL } from '../../backend/graphql'
import { FilteredConnection } from '../../components/FilteredConnection'
import { Form } from '../../components/Form'
import { PageTitle } from '../../components/PageTitle'
import { Timestamp } from '../../components/time/Timestamp'
import { PersonLink } from '../../person/PersonLink'
import { quoteIfNeeded, searchQueryForRepoRev, PatternTypeProps } from '../../search'
import { eventLogger } from '../../tracking/eventLogger'
import { UserAvatar } from '../../user/UserAvatar'
import { RepositoryStatsAreaPageProps } from './RepositoryStatsArea'

interface QuerySpec {
    revisionRange: string | null
    after: string | null
    path: string | null
}

interface RepositoryContributorNodeProps extends QuerySpec, Omit<PatternTypeProps, 'setPatternType'> {
    node: GQL.IRepositoryContributor
    repoName: string
}

const RepositoryContributorNode: React.FunctionComponent<RepositoryContributorNodeProps> = ({
    node,
    repoName,
    revisionRange,
    after,
    path,
    patternType,
}) => {
    const commit = node.commits.nodes[0] as GQL.IGitCommit | undefined

    const query: string = [
        searchQueryForRepoRev(repoName),
        'type:diff',
        `author:${quoteIfNeeded(node.person.email)}`,
        after ? `after:${quoteIfNeeded(after)}` : '',
        path ? `file:${quoteIfNeeded(escapeRegExp(path))}` : '',
    ]
        .join(' ')
        .replace(/\s+/, ' ')

    return (
        <div className="repository-contributor-node list-group-item py-2">
            <div className="repository-contributor-node__person">
                <UserAvatar className="icon-inline mr-2" user={node.person} />
                <PersonLink userClassName="font-weight-bold" person={node.person} />
            </div>
            <div className="repository-contributor-node__commits">
                <div className="repository-contributor-node__commit">
                    {commit && (
                        <>
                            <Timestamp date={commit.author.date} />:{' '}
                            <Link
                                to={commit.url}
                                className="repository-contributor-node__commit-subject"
                                data-tooltip="Most recent commit by contributor"
                            >
                                {commit.subject}
                            </Link>
                        </>
                    )}
                </div>
                <div className="repository-contributor-node__count">
                    <Link
                        to={`/search?${buildSearchURLQuery(query, patternType)}`}
                        className="font-weight-bold"
                        data-tooltip={
                            revisionRange?.includes('..') &&
                            'All commits will be shown (revision end ranges are not yet supported)'
                        }
                    >
                        {numberWithCommas(node.count)} {pluralize('commit', node.count)}
                    </Link>
                </div>
            </div>
        </div>
    )
}

const queryRepositoryContributors = memoizeObservable(
    (args: {
        repo: GQL.ID
        first?: number
        revisionRange?: string
        after?: string
        path?: string
    }): Observable<GQL.IRepositoryContributorConnection> =>
        queryGraphQL(
            gql`
                query RepositoryContributors(
                    $repo: ID!
                    $first: Int
                    $revisionRange: String
                    $after: String
                    $path: String
                ) {
                    node(id: $repo) {
                        ... on Repository {
                            contributors(first: $first, revisionRange: $revisionRange, after: $after, path: $path) {
                                nodes {
                                    person {
                                        name
                                        displayName
                                        email
                                        avatarURL
                                        user {
                                            username
                                            url
                                        }
                                    }
                                    count
                                    commits(first: 1) {
                                        nodes {
                                            oid
                                            abbreviatedOID
                                            url
                                            subject
                                            author {
                                                date
                                            }
                                        }
                                    }
                                }
                                totalCount
                                pageInfo {
                                    hasNextPage
                                }
                            }
                        }
                    }
                }
            `,
            args
        ).pipe(
            map(({ data, errors }) => {
                if (!data || !data.node || !(data.node as GQL.IRepository).contributors || errors) {
                    throw createAggregateError(errors)
                }
                return (data.node as GQL.IRepository).contributors
            })
        ),
    args => `${args.repo}:${args.first}:${args.revisionRange}:${args.after}:${args.path}`
)

interface Props
    extends RepositoryStatsAreaPageProps,
        RouteComponentProps<{}>,
        Omit<PatternTypeProps, 'setPatternType'> {}

class FilteredContributorsConnection extends FilteredConnection<
    GQL.IRepositoryContributor,
    Pick<RepositoryContributorNodeProps, 'repoName' | 'revisionRange' | 'after' | 'path' | 'patternType'>
> {}

interface State extends QuerySpec {}

/** A page that shows a repository's contributors. */
export class RepositoryStatsContributorsPage extends React.PureComponent<Props, State> {
    private static REVISION_RANGE_INPUT_ID = 'repository-stats-contributors-page__revision-range'
    private static AFTER_INPUT_ID = 'repository-stats-contributors-page__after'
    private static PATH_INPUT_ID = 'repository-stats-contributors-page__path'

    public state: State = this.getDerivedProps()

    private specChanges = new Subject<void>()

    public componentDidMount(): void {
        eventLogger.logViewEvent('RepositoryStatsContributors')
    }

    public componentDidUpdate(prevProps: Props): void {
        const spec = this.getDerivedProps(this.props.location)
        const prevSpec = this.getDerivedProps(prevProps.location)
        if (!isEqual(spec, prevSpec)) {
            eventLogger.log('RepositoryStatsContributorsPropsUpdated')
            // eslint-disable-next-line react/no-did-update-set-state
            this.setState(spec)
            this.specChanges.next()
        }
    }

    private getDerivedProps(location = this.props.location): QuerySpec {
        const query = new URLSearchParams(location.search)
        const revisionRange = query.get('revision-range')
        const after = query.get('after')
        const path = query.get('path')
        return { revisionRange, after, path }
    }

    public render(): JSX.Element | null {
        const { revisionRange, after, path } = this.getDerivedProps()

        // Whether the user has entered new option values that differ from what's in the URL query and has not yet
        // submitted the form.
        const equalOrEmpty = (a: string | null, b: string | null): boolean => a === b || (!a && !b)
        const stateDiffers =
            !equalOrEmpty(that.state.revisionRange, revisionRange) ||
            !equalOrEmpty(that.state.after, after) ||
            !equalOrEmpty(that.state.path, path)

        return (
            <div className="repository-stats-page">
                <PageTitle title="Contributors" />
                <div className="card repository-stats-page__card repository-stats-page__card--form">
                    <div className="card-header">Contributions filter</div>
                    <div className="card-body">
                        <Form onSubmit={that.onSubmit}>
                            <div className="repository-stats-page__row form-inline">
                                <div className="input-group mb-2 mr-sm-2">
                                    <div className="input-group-prepend">
                                        <label
                                            htmlFor={RepositoryStatsContributorsPage.AFTER_INPUT_ID}
                                            className="input-group-text"
                                        >
                                            Time period
                                        </label>
                                    </div>
                                    <input
                                        type="text"
                                        className="form-control"
                                        name="after"
                                        size={12}
                                        id={RepositoryStatsContributorsPage.AFTER_INPUT_ID}
                                        value={that.state.after || ''}
                                        placeholder="All time"
                                        onChange={that.onChange}
                                    />
                                    <div className="input-group-append">
                                        <div className="btn-group">
                                            <button
                                                type="button"
                                                className={`btn btn-secondary ${
                                                    that.state.after === '7 days ago' ? 'active' : ''
                                                } repository-stats-page__btn-no-left-rounded-corners`}
                                                onClick={() => that.setStateAfterAndSubmit('7 days ago')}
                                            >
                                                Last 7 days
                                            </button>
                                            <button
                                                type="button"
                                                className={`btn btn-secondary ${
                                                    that.state.after === '30 days ago' ? 'active' : ''
                                                }`}
                                                onClick={() => that.setStateAfterAndSubmit('30 days ago')}
                                            >
                                                Last 30 days
                                            </button>
                                            <button
                                                type="button"
                                                className={`btn btn-secondary ${
                                                    that.state.after === '1 year ago' ? 'active' : ''
                                                }`}
                                                onClick={() => that.setStateAfterAndSubmit('1 year ago')}
                                            >
                                                Last year
                                            </button>
                                            <button
                                                type="button"
                                                className={`btn btn-secondary ${!that.state.after ? 'active' : ''}`}
                                                onClick={() => that.setStateAfterAndSubmit(null)}
                                            >
                                                All time
                                            </button>
                                        </div>
                                    </div>
                                </div>
                            </div>
                            <div className="repository-stats-page__row form-inline">
                                <div className="input-group mt-2 mr-sm-2">
                                    <div className="input-group-prepend">
                                        <label
                                            htmlFor={RepositoryStatsContributorsPage.REVISION_RANGE_INPUT_ID}
                                            className="input-group-text"
                                        >
                                            Revision range
                                        </label>
                                    </div>
                                    <input
                                        type="text"
                                        className="form-control"
                                        name="revision-range"
                                        size={18}
                                        id={RepositoryStatsContributorsPage.REVISION_RANGE_INPUT_ID}
                                        value={that.state.revisionRange || ''}
                                        placeholder="Default branch"
                                        onChange={that.onChange}
                                        autoCapitalize="off"
                                        autoCorrect="off"
                                        autoComplete="off"
                                        spellCheck={false}
                                    />
                                </div>
                                <div className="input-group mt-2 mr-sm-2">
                                    <div className="input-group-prepend">
                                        <label
                                            htmlFor={RepositoryStatsContributorsPage.PATH_INPUT_ID}
                                            className="input-group-text"
                                        >
                                            Path
                                        </label>
                                    </div>
                                    <input
                                        type="text"
                                        className="form-control"
                                        name="path"
                                        size={18}
                                        id={RepositoryStatsContributorsPage.PATH_INPUT_ID}
                                        value={that.state.path || ''}
                                        placeholder="All files"
                                        onChange={that.onChange}
                                        autoCapitalize="off"
                                        autoCorrect="off"
                                        autoComplete="off"
                                        spellCheck={false}
                                    />
                                </div>
                                {stateDiffers && (
                                    <div className="form-group mb-0">
                                        <button type="submit" className="btn btn-primary mr-2 mt-2">
                                            Update
                                        </button>
                                        <button type="reset" className="btn btn-secondary mt-2" onClick={that.onCancel}>
                                            Cancel
                                        </button>
                                    </div>
                                )}
                            </div>
                        </Form>
                    </div>
                </div>
                <FilteredContributorsConnection
                    listClassName="list-group list-group-flush"
                    noun="contributor"
                    pluralNoun="contributors"
                    queryConnection={that.queryRepositoryContributors}
                    nodeComponent={RepositoryContributorNode}
                    nodeComponentProps={{
                        repoName: that.props.repo.name,
                        revisionRange,
                        after,
                        path,
                        patternType: that.props.patternType,
                    }}
                    defaultFirst={20}
                    hideSearch={true}
                    useURLQuery={false}
                    updates={that.specChanges}
                    history={that.props.history}
                    location={that.props.location}
                />
            </div>
        )
    }

    private onChange: React.ChangeEventHandler<HTMLInputElement> = e => {
        switch (e.currentTarget.id) {
            case RepositoryStatsContributorsPage.REVISION_RANGE_INPUT_ID:
                that.setState({ revisionRange: e.currentTarget.value })
                return

            case RepositoryStatsContributorsPage.AFTER_INPUT_ID:
                that.setState({ after: e.currentTarget.value })
                return

            case RepositoryStatsContributorsPage.PATH_INPUT_ID:
                that.setState({ path: e.currentTarget.value })
                return
        }
    }

    private onSubmit: React.FormEventHandler<HTMLFormElement> = e => {
        e.preventDefault()
        that.props.history.push({ search: that.urlQuery(that.state) })
    }

    private onCancel: React.MouseEventHandler<HTMLButtonElement> = e => {
        e.preventDefault()
        that.setState(that.getDerivedProps())
    }

    private queryRepositoryContributors = (args: {
        first?: number
    }): Observable<GQL.IRepositoryContributorConnection> => {
        const { revisionRange, after, path } = that.getDerivedProps()
        return queryRepositoryContributors({
            ...args,
            repo: that.props.repo.id,
            revisionRange: revisionRange || undefined,
            after: after || undefined,
            path: path || undefined,
        })
    }

    private urlQuery(spec: QuerySpec): string {
        const query = new URLSearchParams()
        if (spec.revisionRange) {
            query.set('revision-range', spec.revisionRange)
        }
        if (spec.after) {
            query.set('after', spec.after)
        }
        if (spec.path) {
            query.set('path', spec.path)
        }
        return query.toString()
    }

    private setStateAfterAndSubmit(after: string | null): void {
        that.setState({ after }, () => that.props.history.push({ search: that.urlQuery(that.state) }))
    }
}
