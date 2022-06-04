import * as React from 'react'

import classNames from 'classnames'
import { escapeRegExp, isEqual } from 'lodash'
import { RouteComponentProps } from 'react-router-dom'
import { Observable, Subject } from 'rxjs'
import { map } from 'rxjs/operators'

import { Form } from '@sourcegraph/branded/src/components/Form'
import { createAggregateError, numberWithCommas, pluralize, memoizeObservable } from '@sourcegraph/common'
import { gql } from '@sourcegraph/http-client'
import { Scalars, SearchPatternType } from '@sourcegraph/shared/src/graphql-operations'
import * as GQL from '@sourcegraph/shared/src/schema'
import { buildSearchURLQuery } from '@sourcegraph/shared/src/util/url'
import { Button, ButtonGroup, Link, CardHeader, CardBody, Card, Input, Label } from '@sourcegraph/wildcard'

import { queryGraphQL } from '../../backend/graphql'
import { FilteredConnection } from '../../components/FilteredConnection'
import { PageTitle } from '../../components/PageTitle'
import { Timestamp } from '../../components/time/Timestamp'
import { PersonLink } from '../../person/PersonLink'
import { quoteIfNeeded, searchQueryForRepoRevision } from '../../search'
import { eventLogger } from '../../tracking/eventLogger'
import { UserAvatar } from '../../user/UserAvatar'

import { RepositoryStatsAreaPageProps } from './RepositoryStatsArea'

import styles from './RepositoryStatsContributorsPage.module.scss'

interface QuerySpec {
    revisionRange: string | null
    after: string | null
    path: string | null
}

interface RepositoryContributorNodeProps extends QuerySpec {
    node: GQL.IRepositoryContributor
    repoName: string
    globbing: boolean
}

const RepositoryContributorNode: React.FunctionComponent<React.PropsWithChildren<RepositoryContributorNodeProps>> = ({
    node,
    repoName,
    revisionRange,
    after,
    path,
    globbing,
}) => {
    const commit = node.commits.nodes[0] as GQL.IGitCommit | undefined

    const query: string = [
        searchQueryForRepoRevision(repoName, globbing),
        'type:diff',
        `author:${quoteIfNeeded(node.person.email)}`,
        after ? `after:${quoteIfNeeded(after)}` : '',
        path ? `file:${quoteIfNeeded(escapeRegExp(path))}` : '',
    ]
        .join(' ')
        .replace(/\s+/, ' ')

    return (
        <li className={classNames('list-group-item py-2', styles.repositoryContributorNode)}>
            <div className={styles.person}>
                <UserAvatar inline={true} className="mr-2" user={node.person} />
                <PersonLink userClassName="font-weight-bold" person={node.person} />
            </div>
            <div className={styles.commits}>
                <div className={styles.commit}>
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
                <div className={styles.count}>
                    <Link
                        to={`/search?${buildSearchURLQuery(query, SearchPatternType.literal, false)}`}
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
        </li>
    )
}

const queryRepositoryContributors = memoizeObservable(
    (args: {
        repo: Scalars['ID']
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
    args =>
        `${args.repo}:${String(args.first)}:${String(args.revisionRange)}:${String(args.after)}:${String(args.path)}`
)

const equalOrEmpty = (a: string | null, b: string | null): boolean => a === b || (!a && !b)

interface Props extends RepositoryStatsAreaPageProps, RouteComponentProps<{}> {
    globbing: boolean
}

class FilteredContributorsConnection extends FilteredConnection<
    GQL.IRepositoryContributor,
    Pick<RepositoryContributorNodeProps, 'repoName' | 'revisionRange' | 'after' | 'path' | 'globbing'>
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

    public componentDidUpdate(previousProps: Props): void {
        const spec = this.getDerivedProps(this.props.location)
        const previousSpec = this.getDerivedProps(previousProps.location)
        if (!isEqual(spec, previousSpec)) {
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
        const stateDiffers =
            !equalOrEmpty(this.state.revisionRange, revisionRange) ||
            !equalOrEmpty(this.state.after, after) ||
            !equalOrEmpty(this.state.path, path)

        return (
            <div>
                <PageTitle title="Contributors" />
                <Card className={styles.card}>
                    <CardHeader>Contributions filter</CardHeader>
                    <CardBody>
                        <Form onSubmit={this.onSubmit}>
                            <div className={classNames(styles.row, 'form-inline')}>
                                <div className="input-group mb-2 mr-sm-2">
                                    <div className="input-group-prepend">
                                        <Label
                                            htmlFor={RepositoryStatsContributorsPage.AFTER_INPUT_ID}
                                            className="input-group-text"
                                        >
                                            Time period
                                        </Label>
                                    </div>
                                    <Input
                                        name="after"
                                        size={12}
                                        id={RepositoryStatsContributorsPage.AFTER_INPUT_ID}
                                        value={this.state.after || ''}
                                        placeholder="All time"
                                        onChange={this.onChange}
                                    />
                                    <div className="input-group-append">
                                        <ButtonGroup aria-label="Time period presets">
                                            <Button
                                                className={classNames(
                                                    styles.btnNoLeftRoundedCorners,
                                                    this.state.after === '7 days ago' && 'active'
                                                )}
                                                onClick={() => this.setStateAfterAndSubmit('7 days ago')}
                                                variant="secondary"
                                            >
                                                Last 7 days
                                            </Button>
                                            <Button
                                                className={classNames(this.state.after === '30 days ago' && 'active')}
                                                onClick={() => this.setStateAfterAndSubmit('30 days ago')}
                                                variant="secondary"
                                            >
                                                Last 30 days
                                            </Button>
                                            <Button
                                                className={classNames(this.state.after === '1 year ago' && 'active')}
                                                onClick={() => this.setStateAfterAndSubmit('1 year ago')}
                                                variant="secondary"
                                            >
                                                Last year
                                            </Button>
                                            <Button
                                                className={classNames(!this.state.after && 'active')}
                                                onClick={() => this.setStateAfterAndSubmit(null)}
                                                variant="secondary"
                                            >
                                                All time
                                            </Button>
                                        </ButtonGroup>
                                    </div>
                                </div>
                            </div>
                            <div className={classNames(styles.row, 'form-inline')}>
                                <div className="input-group mt-2 mr-sm-2">
                                    <div className="input-group-prepend">
                                        <Label
                                            htmlFor={RepositoryStatsContributorsPage.REVISION_RANGE_INPUT_ID}
                                            className="input-group-text"
                                        >
                                            Revision range
                                        </Label>
                                    </div>
                                    <Input
                                        name="revision-range"
                                        size={18}
                                        id={RepositoryStatsContributorsPage.REVISION_RANGE_INPUT_ID}
                                        value={this.state.revisionRange || ''}
                                        placeholder="Default branch"
                                        onChange={this.onChange}
                                        autoCapitalize="off"
                                        autoCorrect="off"
                                        autoComplete="off"
                                        spellCheck={false}
                                    />
                                </div>
                                <div className="input-group mt-2 mr-sm-2">
                                    <div className="input-group-prepend">
                                        <Label
                                            htmlFor={RepositoryStatsContributorsPage.PATH_INPUT_ID}
                                            className="input-group-text"
                                        >
                                            Path
                                        </Label>
                                    </div>
                                    <Input
                                        name="path"
                                        size={18}
                                        id={RepositoryStatsContributorsPage.PATH_INPUT_ID}
                                        value={this.state.path || ''}
                                        placeholder="All files"
                                        onChange={this.onChange}
                                        autoCapitalize="off"
                                        autoCorrect="off"
                                        autoComplete="off"
                                        spellCheck={false}
                                    />
                                </div>
                                {stateDiffers && (
                                    <div className="form-group mb-0">
                                        <Button type="submit" className="mr-2 mt-2" variant="primary">
                                            Update
                                        </Button>
                                        <Button
                                            type="reset"
                                            className="mt-2"
                                            onClick={this.onCancel}
                                            variant="secondary"
                                        >
                                            Cancel
                                        </Button>
                                    </div>
                                )}
                            </div>
                        </Form>
                    </CardBody>
                </Card>
                <FilteredContributorsConnection
                    listClassName="list-group list-group-flush test-filtered-contributors-connection"
                    noun="contributor"
                    pluralNoun="contributors"
                    queryConnection={this.queryRepositoryContributors}
                    nodeComponent={RepositoryContributorNode}
                    nodeComponentProps={{
                        repoName: this.props.repo.name,
                        revisionRange,
                        after,
                        path,
                        globbing: this.props.globbing,
                    }}
                    defaultFirst={20}
                    hideSearch={true}
                    useURLQuery={false}
                    updates={this.specChanges}
                    history={this.props.history}
                    location={this.props.location}
                />
            </div>
        )
    }

    private onChange: React.ChangeEventHandler<HTMLInputElement> = event => {
        switch (event.currentTarget.id) {
            case RepositoryStatsContributorsPage.REVISION_RANGE_INPUT_ID:
                this.setState({ revisionRange: event.currentTarget.value })
                return

            case RepositoryStatsContributorsPage.AFTER_INPUT_ID:
                this.setState({ after: event.currentTarget.value })
                return

            case RepositoryStatsContributorsPage.PATH_INPUT_ID:
                this.setState({ path: event.currentTarget.value })
                return
        }
    }

    private onSubmit: React.FormEventHandler<HTMLFormElement> = event => {
        event.preventDefault()
        this.props.history.push({ search: this.urlQuery(this.state) })
    }

    private onCancel: React.MouseEventHandler<HTMLButtonElement> = event => {
        event.preventDefault()
        this.setState(this.getDerivedProps())
    }

    private queryRepositoryContributors = (args: {
        first?: number
    }): Observable<GQL.IRepositoryContributorConnection> => {
        const { revisionRange, after, path } = this.getDerivedProps()
        return queryRepositoryContributors({
            ...args,
            repo: this.props.repo.id,
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
        this.setState({ after }, () => this.props.history.push({ search: this.urlQuery(this.state) }))
    }
}
