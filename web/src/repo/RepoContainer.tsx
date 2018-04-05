import DirectionalSignIcon from '@sourcegraph/icons/lib/DirectionalSign'
import ErrorIcon from '@sourcegraph/icons/lib/Error'
import escapeRegexp from 'escape-string-regexp'
import * as React from 'react'
import { Route, RouteComponentProps, Switch } from 'react-router'
import { merge } from 'rxjs/observable/merge'
import { catchError } from 'rxjs/operators/catchError'
import { distinctUntilChanged } from 'rxjs/operators/distinctUntilChanged'
import { map } from 'rxjs/operators/map'
import { switchMap } from 'rxjs/operators/switchMap'
import { tap } from 'rxjs/operators/tap'
import { withLatestFrom } from 'rxjs/operators/withLatestFrom'
import { Subject } from 'rxjs/Subject'
import { Subscription } from 'rxjs/Subscription'
import { parseBrowserRepoURL } from '.'
import { parseRepoRev, redirectToExternalHost } from '.'
import { HeroPage } from '../components/HeroPage'
import { queryUpdates } from '../search/QueryInput'
import { ErrorLike, isErrorLike } from '../util/errors'
import { GoToCodeHostAction } from './actions/GoToCodeHostAction'
import { EREPONOTFOUND, EREPOSEEOTHER, fetchRepository, RepoSeeOtherError } from './backend'
import { RepositoryBranchesArea } from './branches/RepositoryBranchesArea'
import { RepositoryCompareArea } from './compare/RepositoryCompareArea'
import { RepositoryGraphAction } from './graph/RepositoryGraphAction'
import { RepositoryReleasesArea } from './releases/RepositoryReleasesArea'
import { RepoHeader } from './RepoHeader'
import { RepoHeaderActionPortal } from './RepoHeaderActionPortal'
import { RepoRevContainer } from './RepoRevContainer'
import { RepositoryErrorPage } from './RepositoryErrorPage'
import { RepoSettingsArea } from './settings/RepoSettingsArea'

const RepoPageNotFound: React.SFC = () => (
    <HeroPage icon={DirectionalSignIcon} title="404: Not Found" subtitle="The repository page was not found." />
)

interface Props extends RouteComponentProps<{ repoRevAndRest: string }> {
    user: GQL.IUser | null
    isLightTheme: boolean
}

interface State {
    repoPath: string
    rev?: string
    filePath?: string
    rest?: string

    repoHeaderLeftChildren?: React.ReactFragment | null
    repoHeaderRightChildren?: React.ReactFragment | null

    /**
     * The fetched repository or an error if occured.
     * `undefined` while loading.
     */
    repoOrError?: GQL.IRepository | ErrorLike
}

const enableRepositoryGraph = localStorage.getItem('repositoryGraph') !== null

/**
 * Renders a horizontal bar and content for a repository page.
 */
export class RepoContainer extends React.Component<Props, State> {
    private routeMatchChanges = new Subject<{ repoRevAndRest: string }>()
    private repositoryUpdates = new Subject<Partial<GQL.IRepository>>()
    private repositoryAdds = new Subject<void>()
    private subscriptions = new Subscription()

    constructor(props: Props) {
        super(props)

        this.state = {
            ...parseURLPath(props.match.params.repoRevAndRest),
        }
    }

    public componentDidMount(): void {
        const parsedRouteChanges = this.routeMatchChanges.pipe(
            map(({ repoRevAndRest }) => parseURLPath(repoRevAndRest))
        )

        // Fetch repository.
        const repositoryChanges = parsedRouteChanges.pipe(map(({ repoPath }) => repoPath), distinctUntilChanged())
        this.subscriptions.add(
            merge(
                repositoryChanges,
                this.repositoryAdds.pipe(withLatestFrom(repositoryChanges), map(([, repoPath]) => repoPath))
            )
                .pipe(
                    tap(() => this.setState({ repoOrError: undefined })),
                    switchMap(repoPath =>
                        fetchRepository({ repoPath }).pipe(
                            catchError(error => {
                                switch (error.code) {
                                    case EREPOSEEOTHER:
                                        redirectToExternalHost((error as RepoSeeOtherError).redirectURL)
                                        return []
                                }
                                this.setState({ repoOrError: error })
                                return []
                            })
                        )
                    )
                )
                .subscribe(
                    repo => {
                        this.setState({ repoOrError: repo })
                    },
                    err => {
                        console.error(err)
                    }
                )
        )

        // Update header and other global state.
        this.subscriptions.add(
            parsedRouteChanges.subscribe(({ repoPath, rev, rest }) => {
                this.setState({ repoPath, rev, rest })

                queryUpdates.next(searchQueryForRepoRev(repoPath, rev))
            })
        )

        this.routeMatchChanges.next(this.props.match.params)

        // Merge in repository updates.
        this.subscriptions.add(
            this.repositoryUpdates.subscribe(update =>
                this.setState(({ repoOrError }) => ({ repoOrError: { ...repoOrError, ...update } as GQL.IRepository }))
            )
        )
    }

    public componentWillReceiveProps(props: Props): void {
        if (props.match.params !== this.props.match.params) {
            this.routeMatchChanges.next(props.match.params)
        }
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        if (!this.state.repoOrError) {
            // Render nothing while loading
            return null
        }

        const { repoPath, filePath, position, range } = parseBrowserRepoURL(
            location.pathname + location.search + location.hash
        )
        const viewerCanAdminister = !!this.props.user && this.props.user.siteAdmin

        if (isErrorLike(this.state.repoOrError)) {
            // Display error page
            switch (this.state.repoOrError.code) {
                case EREPONOTFOUND:
                    return (
                        <RepositoryErrorPage
                            repo={repoPath}
                            repoID={null}
                            error={this.state.repoOrError}
                            viewerCanAdminister={viewerCanAdminister}
                            onDidAddRepository={this.onDidAddRepository}
                        />
                    )
                default:
                    return <HeroPage icon={ErrorIcon} title="Error" subtitle={this.state.repoOrError.message} />
            }
        }

        const transferProps = {
            repo: this.state.repoOrError,
            user: this.props.user,
            isLightTheme: this.props.isLightTheme,
        }

        const repoMatchURL = `/${this.state.repoOrError.uri}`

        const isSettingsPage =
            location.pathname === `${repoMatchURL}/-/settings` ||
            location.pathname.startsWith(`${repoMatchURL}/-/settings/`)

        return (
            <div className="repo-composite-container composite-container">
                <RepoHeader
                    repo={this.state.repoOrError}
                    rev={this.state.rev}
                    filePath={filePath}
                    className="repo-composite-container__header"
                    location={this.props.location}
                    history={this.props.history}
                />
                {enableRepositoryGraph && (
                    <RepoHeaderActionPortal
                        position="right"
                        priority={-1}
                        element={
                            <RepositoryGraphAction
                                key="repository-graph"
                                repo={this.state.repoOrError.uri}
                                rev={this.state.rev}
                            />
                        }
                    />
                )}
                <RepoHeaderActionPortal
                    position="right"
                    key="go-to-code-host"
                    element={
                        <GoToCodeHostAction
                            key="go-to-code-host"
                            repo={this.state.repoOrError}
                            // We need a rev to generate code host URLs, since we don't have a default use HEAD.
                            rev={this.state.rev || 'HEAD'}
                            filePath={filePath}
                            position={position}
                            range={range}
                        />
                    }
                />
                {this.state.repoOrError.enabled || isSettingsPage ? (
                    <Switch>
                        {['', `@${this.state.rev}`, '/-/blob', '/-/tree', '/-/graph', '/-/commits'].map(routePath => (
                            <Route
                                path={`${repoMatchURL}${routePath}`}
                                key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                                exact={routePath === ''}
                                // tslint:disable-next-line:jsx-no-lambda
                                render={routeComponentProps => (
                                    <RepoRevContainer
                                        {...routeComponentProps}
                                        {...transferProps}
                                        rev={this.state.rev}
                                        routePrefix={`${repoMatchURL}${this.state.rev ? `@${this.state.rev}` : ''}`}
                                    />
                                )}
                            />
                        ))}
                        <Route
                            path={`${repoMatchURL}/-/branches`}
                            key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                            // tslint:disable-next-line:jsx-no-lambda
                            render={routeComponentProps => (
                                <RepositoryBranchesArea {...routeComponentProps} {...transferProps} />
                            )}
                        />
                        <Route
                            path={`${repoMatchURL}/-/tags`}
                            key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                            // tslint:disable-next-line:jsx-no-lambda
                            render={routeComponentProps => (
                                <RepositoryReleasesArea
                                    {...routeComponentProps}
                                    {...transferProps}
                                    repoMatchURL={repoMatchURL}
                                />
                            )}
                        />
                        <Route
                            path={`${repoMatchURL}/-/compare/:spec*`}
                            key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                            // tslint:disable-next-line:jsx-no-lambda
                            render={routeComponentProps => (
                                <RepositoryCompareArea {...routeComponentProps} {...transferProps} />
                            )}
                        />
                        <Route
                            path={`${repoMatchURL}/-/settings`}
                            key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                            // tslint:disable-next-line:jsx-no-lambda
                            render={routeComponentProps => (
                                <RepoSettingsArea
                                    {...routeComponentProps}
                                    {...transferProps}
                                    onDidUpdateRepository={this.onDidUpdateRepository}
                                />
                            )}
                        />
                        <Route key="hardcoded-key" component={RepoPageNotFound} />
                    </Switch>
                ) : (
                    <RepositoryErrorPage
                        repo={this.state.repoOrError.uri}
                        repoID={this.state.repoOrError.id}
                        error="disabled"
                        viewerCanAdminister={viewerCanAdminister}
                        onDidUpdateRepository={this.onDidUpdateRepository}
                    />
                )}
            </div>
        )
    }

    private onDidUpdateRepository = (update: Partial<GQL.IRepository>) => this.repositoryUpdates.next(update)
    private onDidAddRepository = () => this.repositoryAdds.next()
}

/**
 * Parses the URL path (without the leading slash).
 *
 * TODO(sqs): replace with parseBrowserRepoURL?
 *
 * @param repoRevAndRest a string like /my/repo@myrev/-/blob/my/file.txt
 */
function parseURLPath(repoRevAndRest: string): { repoPath: string; rev?: string; rest?: string } {
    const [repoRev, rest] = repoRevAndRest.split('/-/', 2)
    const { repoPath, rev } = parseRepoRev(repoRev)
    return { repoPath, rev, rest }
}

function abbreviateOID(oid: string): string {
    if (oid.length === 40) {
        return oid.slice(0, 7)
    }
    return oid
}

export function searchQueryForRepoRev(repoPath: string, rev?: string): string {
    return `repo:^${escapeRegexp(repoPath)}$${rev ? `@${abbreviateOID(rev)}` : ''} `
}
