import DirectionalSignIcon from '@sourcegraph/icons/lib/DirectionalSign'
import escapeRegexp from 'escape-string-regexp'
import * as React from 'react'
import { Route, RouteComponentProps, Switch } from 'react-router'
import { catchError } from 'rxjs/operators/catchError'
import { distinctUntilChanged } from 'rxjs/operators/distinctUntilChanged'
import { map } from 'rxjs/operators/map'
import { switchMap } from 'rxjs/operators/switchMap'
import { tap } from 'rxjs/operators/tap'
import { Subject } from 'rxjs/Subject'
import { Subscription } from 'rxjs/Subscription'
import { parseRepoRev } from '.'
import { HeroPage } from '../components/HeroPage'
import { queryUpdates } from '../search/QueryInput'
import { RouteWithProps } from '../util/RouteWithProps'
import { GoToGitHubAction } from './actions/GoToGitHubAction'
import { GoToPhabricatorAction } from './actions/GoToPhabricator'
import { fetchRepository } from './backend'
import { RepoHeader } from './RepoHeader'
import { RepoHeaderActionPortal } from './RepoHeaderActionPortal'
import { RepoRevContainer } from './RepoRevContainer'
import { RepoSettingsArea } from './settings/RepoSettingsArea'

const RepoPageNotFound: React.SFC = () => (
    <HeroPage icon={DirectionalSignIcon} title="404: Not Found" subtitle="The repository page was not found." />
)

interface Props extends RouteComponentProps<{ repoRevAndRest: string }> {
    user: GQL.IUser | null
}

interface State {
    repoPath: string
    rev?: string
    rest?: string

    loading: boolean

    repoHeaderLeftChildren?: React.ReactFragment | null
    repoHeaderRightChildren?: React.ReactFragment | null

    repo?: GQL.IRepository | null
    error?: string
}

/**
 * Renders a horizontal bar and content for a repository page.
 */
export class RepoContainer extends React.Component<Props, State> {
    private routeMatchChanges = new Subject<{ repoRevAndRest: string }>()
    private subscriptions = new Subscription()

    constructor(props: Props) {
        super(props)

        this.state = {
            ...parseURLPath(props.match.params.repoRevAndRest),
            loading: true,
        }
    }

    public componentDidMount(): void {
        const parsedRouteChanges = this.routeMatchChanges.pipe(
            map(({ repoRevAndRest }) => parseURLPath(repoRevAndRest))
        )

        // Fetch repository.
        this.subscriptions.add(
            parsedRouteChanges
                .pipe(
                    map(({ repoPath }) => repoPath),
                    distinctUntilChanged(),
                    tap(() => this.setState({ loading: true })),
                    switchMap(repoPath =>
                        fetchRepository({ repoPath }).pipe(
                            catchError(error => {
                                console.error(error)
                                this.setState({ loading: false, error: error.message })
                                return []
                            })
                        )
                    )
                )
                .subscribe(repo => this.setState({ repo, loading: false }), err => console.error(err))
        )

        // Update header and other global state.
        this.subscriptions.add(
            parsedRouteChanges.subscribe(({ repoPath, rev, rest }) => {
                this.setState({ repoPath, rev, rest })

                queryUpdates.next(searchQueryForRepoRev(repoPath, rev))
            })
        )

        this.routeMatchChanges.next(this.props.match.params)
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
        if (this.state.loading) {
            return null
        }

        if (this.state.error) {
            return <HeroPage icon={DirectionalSignIcon} title="Error" subtitle={this.state.error} />
        }

        if (this.state.repo === undefined) {
            return null
        }
        if (this.state.repo === null) {
            return (
                <HeroPage icon={DirectionalSignIcon} title="404: Not Found" subtitle="The repository was not found." />
            )
        }

        const transferProps = {
            repo: this.state.repo,
            user: this.props.user,
        }

        const repoMatchURL = `/${this.state.repo.uri}`

        return (
            <div className="repo-composite-container composite-container">
                <RepoHeader
                    repo={this.state.repo}
                    rev={this.state.rev}
                    filePath={this.state.rest ? extractFilePathFromRest(this.state.rest) : undefined}
                    className="repo-composite-container__header"
                    location={this.props.location}
                    history={this.props.history}
                />
                <RepoHeaderActionPortal
                    position="right"
                    key="go-to-github"
                    element={<GoToGitHubAction key="go-to-github" location={this.props.location} />}
                />
                <RepoHeaderActionPortal
                    position="right"
                    key="go-to-phabricator"
                    element={
                        <GoToPhabricatorAction
                            key="go-to-phabricator"
                            repo={this.state.repoPath}
                            location={this.props.location}
                        />
                    }
                />
                <Switch>
                    <RouteWithProps
                        path={`${repoMatchURL}`}
                        component={RepoRevContainer}
                        exact={true}
                        other={{ ...transferProps, rev: this.state.rev, objectType: 'tree' }}
                    />
                    <RouteWithProps
                        path={`${repoMatchURL}/-/blob/:filePath+`}
                        component={RepoRevContainer}
                        exact={true}
                        other={{ ...transferProps, rev: this.state.rev, objectType: 'blob' }}
                    />
                    <RouteWithProps
                        path={`${repoMatchURL}@${this.state.rev}/-/blob/:filePath+`}
                        component={RepoRevContainer}
                        exact={true}
                        other={{ ...transferProps, rev: this.state.rev, objectType: 'blob' }}
                    />
                    <RouteWithProps
                        path={`${repoMatchURL}/-/tree/:filePath+`}
                        component={RepoRevContainer}
                        exact={true}
                        other={{ ...transferProps, rev: this.state.rev, objectType: 'tree' }}
                    />
                    <RouteWithProps
                        path={`${repoMatchURL}@${this.state.rev}/-/tree/:filePath+`}
                        component={RepoRevContainer}
                        exact={true}
                        other={{ ...transferProps, rev: this.state.rev, objectType: 'tree' }}
                    />
                    <RouteWithProps
                        path={`${repoMatchURL}@${this.state.rev}`}
                        component={RepoRevContainer}
                        exact={true}
                        other={{ ...transferProps, rev: this.state.rev, objectType: 'tree' }}
                    />
                    <RouteWithProps
                        path={`${repoMatchURL}/-/settings`}
                        component={RepoSettingsArea}
                        exact={true}
                        other={transferProps}
                    />
                    <Route component={RepoPageNotFound} />
                </Switch>
            </div>
        )
    }
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

function extractFilePathFromRest(rest: string): string | undefined {
    if (rest.startsWith('blob/')) {
        return rest.slice('blob/'.length)
    }
    if (rest.startsWith('tree/')) {
        return rest.slice('tree/'.length)
    }
    return undefined
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
