
// Polyfill URL because Chrome and Firefox are not spec-compliant
// Hostnames of URIs with custom schemes (e.g. git) are not parsed out
import { URL, URLSearchParams } from 'whatwg-url'
Object.assign(window, { URL, URLSearchParams })

import ServerIcon from '@sourcegraph/icons/lib/Server'
import * as React from 'react'
import { render } from 'react-dom'
import { BrowserRouter, Route, RouteComponentProps, Switch } from 'react-router-dom'
import 'rxjs/add/observable/defer'
import 'rxjs/add/operator/catch'
import 'rxjs/add/operator/delay'
import 'rxjs/add/operator/do'
import 'rxjs/add/operator/retryWhen'
import 'rxjs/add/operator/switchMap'
import { Observable } from 'rxjs/Observable'
import { Subject } from 'rxjs/Subject'
import { Subscription } from 'rxjs/Subscription'
import { HeroPage } from 'sourcegraph/components/HeroPage'
import { Home } from 'sourcegraph/home/Home'
import { Navbar } from 'sourcegraph/nav/Navbar'
import { ECLONEINPROGESS, EREPONOTFOUND, resolveRev } from 'sourcegraph/repo/backend'
import { Repository, RepositoryCloneInProgress, RepositoryNotFound } from 'sourcegraph/repo/Repository'
import { SearchResults } from 'sourcegraph/search/SearchResults'
import { EditorAuthPage } from 'sourcegraph/user/EditorAuthPage'
import { SignInPage } from 'sourcegraph/user/SignInPage'
import { ParsedRouteProps, parseRouteProps } from 'sourcegraph/util/routes'
import { sourcegraphContext } from 'sourcegraph/util/sourcegraphContext'

interface WithResolvedRevProps {
    component: any
    cloningComponent?: any
    notFoundComponent?: any // for 404s
    repoPath?: string
    rev?: string
    [key: string]: any
}

interface WithResolvedRevState {
    commitID?: string
    cloneInProgress: boolean
    notFound: boolean
}

class WithResolvedRev extends React.Component<WithResolvedRevProps, WithResolvedRevState> {
    public state: WithResolvedRevState = { cloneInProgress: false, notFound: false }
    private componentUpdates = new Subject<WithResolvedRevProps>()
    private subscriptions = new Subscription()

    constructor(props: WithResolvedRevProps) {
        super(props)
        this.subscriptions.add(
            this.componentUpdates
                .switchMap(({ repoPath, rev }) => {
                    if (!repoPath) {
                        return [undefined]
                    }
                    // Defer Observable so it retries the request on resubscription
                    return Observable.defer(() => resolveRev({ repoPath, rev }))
                        // On a CloneInProgress error, retry after 5s
                        .retryWhen(errors => errors
                            .do(err => {
                                if (err.code === ECLONEINPROGESS) {
                                    // Display cloning screen to the user and retry
                                    this.setState({ cloneInProgress: true })
                                    return
                                }
                                if (err.code === EREPONOTFOUND) {
                                    // Display 404to the user and do not retry
                                    this.setState({ notFound: true })
                                }
                                // Don't retry other errors
                                throw err
                            })
                            .delay(1000)
                        )
                        // Log other errors but don't break the stream
                        .catch(err => {
                            console.error(err)
                            return []
                        })
                })
                .subscribe(commitID => {
                    this.setState({ commitID, cloneInProgress: false })
                }, err => console.error(err))
        )
    }

    public componentDidMount(): void {
        this.componentUpdates.next(this.props)
    }

    public componentWillReceiveProps(nextProps: WithResolvedRevProps): void {
        if (this.props.repoPath !== nextProps.repoPath || this.props.rev !== nextProps.rev) {
            // clear state so the child won't render until the revision is resolved for new props
            this.state = { cloneInProgress: false, notFound: false }
            this.componentUpdates.next(nextProps)
        }
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        if (this.props.notFoundComponent && this.state.notFound) {
            return <this.props.notFoundComponent {...this.props} />
        }
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
                return <SearchResults {...this.props} />

            case 'sign-in':
                return <SignInPage showEditorFlow={false} />

            case 'editor-auth':
                if (sourcegraphContext.user) {
                    return <EditorAuthPage />
                }
                return <SignInPage showEditorFlow={true} />

            case 'repository':
                return <WithResolvedRev {...this.props} component={Repository} cloningComponent={RepositoryCloneInProgress} notFoundComponent={RepositoryNotFound} />

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
                <WithResolvedRev {...props} component={Navbar} cloningComponent={Navbar} notFoundComponent={Navbar} />
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
        if (window.pageError && window.pageError.StatusCode !== 404) {
            const statusText = window.pageError.StatusText
            const errorMessage = window.pageError.Error
            const errorID = window.pageError.ErrorID

            let subtitle: JSX.Element | undefined
            if (errorID) {
                subtitle = (
                    <p>Sorry, there's been a problem. Please <a href='mailto:support@sourcegraph.com'>contact us</a> and include the error ID:
                        <span className='error-id'>{errorID}</span>
                    </p>
                )
            }
            if (errorMessage) {
                subtitle = (
                    <div className='app__error'>
                        {subtitle}
                        {subtitle && <hr />}
                        <pre>{errorMessage}</pre>
                    </div>
                )
            } else {
                subtitle = <div className='app__error'>{subtitle}</div>
            }
            return <HeroPage icon={ServerIcon} title={'500: ' + statusText} subtitle={subtitle} />
        }

        return (
            <BrowserRouter>
                <Switch>
                    <Route exact={true} path='/' component={Home} />
                    <Route path='/*' component={Layout} />
                </Switch>
            </BrowserRouter>
        )
    }
}

window.addEventListener('DOMContentLoaded', () => {
    render(<App />, document.querySelector('#root'))
})
