import * as React from 'react'
import { render } from 'react-dom'
import { BrowserRouter, Route, RouteComponentProps, Switch } from 'react-router-dom'
import 'rxjs/add/observable/fromPromise'
import 'rxjs/add/operator/catch'
import { Observable } from 'rxjs/Observable'
import { Subject } from 'rxjs/Subject'
import { Subscription } from 'rxjs/Subscription'
import { Home } from 'sourcegraph/home/Home'
import { Navbar } from 'sourcegraph/nav/Navbar'
import { ResolvedRev, resolveRev } from 'sourcegraph/repo/backend'
import { Repository, RepositoryCloneInProgress } from 'sourcegraph/repo/Repository'
import { SearchResults } from 'sourcegraph/search/SearchResults'
import * as activeRepos from 'sourcegraph/util/activeRepos'
import { ParsedRouteProps, parseRouteProps } from 'sourcegraph/util/routes'

window.addEventListener('DOMContentLoaded', () => {
    // Be a bit proactive and try to fetch/store active repos now. This helps
    // on the first search query, and when the data in local storage is stale.
    activeRepos.get().catch(err => console.error(err))
})

interface WithResolvedRevProps {
    component: any
    cloningComponent?: any
    repoPath?: string
    rev?: string
    [key: string]: any
}

interface WithResolvedRevState {
    commitID?: string
    cloneInProgress: boolean
}

class WithResolvedRev extends React.Component<WithResolvedRevProps, WithResolvedRevState> {
    public state: WithResolvedRevState = { cloneInProgress: false }
    private componentUpdates = new Subject<WithResolvedRevProps>()
    private subscriptions = new Subscription()
    private cloneInProgressRefetchTimers: any[] = []

    constructor(props: WithResolvedRevProps) {
        super(props)
        this.subscriptions.add(
            this.componentUpdates
                .switchMap(props => {
                    if (props.repoPath) {
                        return Observable.fromPromise(resolveRev({ repoPath: props.repoPath, rev: props.rev }))
                            .catch(err => {
                                console.error(err)
                                return []
                            })

                    }
                    const resolved: ResolvedRev = { cloneInProgress: false }
                    return [resolved]
                })
                .subscribe(resolved => {
                    if (resolved.cloneInProgress) {
                         // refetch every 5 seconds
                        this.cloneInProgressRefetchTimers.push(setTimeout(() => this.componentUpdates.next(this.props), 5000))
                    }
                    this.setState(resolved)
                }, err => console.error(err))
        )
    }

    public componentDidMount(): void {
        this.componentUpdates.next(this.props)
    }

    public componentWillReceiveProps(nextProps: WithResolvedRevProps): void {
        if (this.props.repoPath !== nextProps.repoPath || this.props.rev !== nextProps.rev) {
            // clear state so the child won't render until the revision is resolved for new props
            this.state = { cloneInProgress: false }
            this.componentUpdates.next(nextProps)
        }
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
        for (const timer of this.cloneInProgressRefetchTimers) {
            clearTimeout(timer)
        }
        this.cloneInProgressRefetchTimers = []
    }

    public render(): JSX.Element | null {
        if (this.props.cloningComponent && this.state.cloneInProgress) {
            return <this.props.cloningComponent {...this.props} />
        }
        if (this.props.repoPath && !this.state.commitID) {
            // commit not yet resolved but required if repoPath prop is provided;
            // render empty until commit resolved
            return null
        }
        return <this.props.component {...this.props} commitID={this.state.commitID} />
    }
}

class AppRouter extends React.Component<ParsedRouteProps, {}> {
    public render(): JSX.Element | null {
        switch (this.props.routeName) {
            case 'search':
                return <SearchResults />

            case 'repository':
                return <WithResolvedRev {...this.props} component={Repository} cloningComponent={RepositoryCloneInProgress} />

            default:
                return null
        }
    }
}

/**
 * Defines the layout of all pages that have a navbar
 */
class Layout extends React.Component<RouteComponentProps<string[]>, {}> {
    public render(): JSX.Element | null {
        const props = parseRouteProps(this.props)
        return (
            <div className='layout'>
                <WithResolvedRev {...props} component={Navbar} cloningComponent={Navbar} />
                <div className='layout__app-router-container'>
                    <AppRouter {...props} />
                </div>
            </div>
        )
    }
}

/**
 * The root component
 */
class App extends React.Component<{}, {}> {
    public render(): JSX.Element | null {
        return <BrowserRouter>
            <Switch>
                <Route exact path='/' component={Home} />
                <Route path='/*' component={Layout} />
            </Switch>
        </BrowserRouter>
    }
}

window.addEventListener('DOMContentLoaded', () => {
    render(<App />, document.querySelector('#root'))
})
