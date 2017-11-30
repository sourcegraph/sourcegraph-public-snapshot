// Polyfill URL because Chrome and Firefox are not spec-compliant
// Hostnames of URIs with custom schemes (e.g. git) are not parsed out
import './util/polyfill'

import ErrorIcon from '@sourcegraph/icons/lib/Error'
import ServerIcon from '@sourcegraph/icons/lib/Server'
import * as React from 'react'
import { render } from 'react-dom'
import { Route, RouteComponentProps, Switch } from 'react-router'
import { BrowserRouter, Redirect } from 'react-router-dom'
import { fetchCurrentUser } from './auth'
import { HeroPage } from './components/HeroPage'
import { updateUserSessionStores } from './marketing/util'
import { Navbar } from './nav/Navbar'
import { routes } from './routes'
import { parseSearchURLQuery } from './search2/index'
import { SearchPage as SearchPage2 } from './search2/SearchPage'
import { eventLogger } from './tracking/eventLogger'

interface LayoutProps extends RouteComponentProps<any> {
    onToggleTheme: () => void
    isLightTheme: boolean
}

/**
 * Defines the layout of all pages that have a navbar
 */
class Layout extends React.Component<LayoutProps> {
    public render(): JSX.Element | null {
        return (
            <div className="layout">
                <Navbar
                    location={this.props.location}
                    history={this.props.history}
                    onToggleTheme={this.props.onToggleTheme}
                    isLightTheme={this.props.isLightTheme}
                />
                <Switch>
                    {routes.map((route, i) => {
                        const isFullWidth = !route.forceNarrowWidth
                        const Component = route.component
                        return (
                            <Route
                                {...route}
                                key="hardcoded-key" // see https://github.com/ReactTraining/react-router/issues/4578#issuecomment-334489490
                                component={undefined}
                                // tslint:disable-next-line:jsx-no-lambda
                                render={props => (
                                    <div
                                        className={[
                                            'layout__app-router-container',
                                            `layout__app-router-container--${
                                                isFullWidth ? 'full-width' : 'restricted'
                                            }`,
                                        ].join(' ')}
                                    >
                                        {Component && (
                                            <Component
                                                {...props}
                                                isFullWidth={isFullWidth}
                                                onToggleTheme={this.props.onToggleTheme}
                                                isLightTheme={this.props.isLightTheme}
                                            />
                                        )}
                                        {route.render &&
                                            route.render({
                                                ...props,
                                                isFullWidth,
                                                isLightTheme: this.props.isLightTheme,
                                            })}
                                    </div>
                                )}
                            />
                        )
                    })}
                </Switch>
            </div>
        )
    }
}

interface SearchRouterProps extends RouteComponentProps<{}> {
    onToggleTheme: () => void
    isLightTheme: boolean
}

/**
 * handles rendering Search or SearchResults components based on whether or not
 * the search query (e.g. '?q=foo') is in URL.
 */
const SearchRouter = (props: SearchRouterProps): JSX.Element | null => {
    const options = parseSearchURLQuery(props.location.search)
    if (options && options.query) {
        return <Layout {...props} onToggleTheme={props.onToggleTheme} isLightTheme={props.isLightTheme} />
    }
    return <SearchPage2 {...props} onToggleTheme={props.onToggleTheme} isLightTheme={props.isLightTheme} />
}

interface BackfillRedirectorProps extends RouteComponentProps<{}> {
    onToggleTheme: () => void
    isLightTheme: boolean
}

interface BackfillRedirectorProps extends RouteComponentProps<{}> {
    onToggleTheme: () => void
    isLightTheme: boolean
}

/**
 * handles rendering Search or SearchResults components based on whether or not
 * the search query (e.g. '?q=foo') is in URL.
 */

class BackfillRedirector extends React.Component<BackfillRedirectorProps, { returnTo: string }> {
    constructor(props: BackfillRedirectorProps) {
        super(props)
        const searchParams = new URLSearchParams(this.props.location.search)
        this.state = {
            returnTo: searchParams.get('returnTo') || window.location.href,
        }
    }

    private renderSearchRouter = (props: RouteComponentProps<any>) => <SearchRouter {...this.props} {...props} />
    private renderLayout = (props: RouteComponentProps<any>) => <Layout {...this.props} {...props} />

    public render(): JSX.Element {
        const searchParams = new URLSearchParams(this.props.location.search)

        const redirectToBackfill =
            window.context.user &&
            window.context.requireUserBackfill &&
            this.props.location.pathname !== '/settings' &&
            searchParams.get('backfill') !== 'true'

        if (redirectToBackfill) {
            return <Redirect to={`/settings?backfill=true&returnTo=${encodeURIComponent(this.state.returnTo)}`} />
        }
        return (
            <Switch>
                <Route path="/search" exact={true} render={this.renderSearchRouter} />
                <Route render={this.renderLayout} />
            </Switch>
        )
    }
}

interface AppState {
    error?: Error
    /**
     * whether or not container is light themed
     */
    isLightTheme: boolean
}

/**
 * The root component
 */
class App extends React.Component<{}, AppState> {
    public state: AppState = {
        isLightTheme: localStorage.getItem('light-theme') === 'true',
    }

    constructor(props: {}) {
        super(props)
        // Fetch current user data
        fetchCurrentUser(this.state.isLightTheme).subscribe(undefined, error => {
            console.error(error)
            this.setState({ error })
        })
    }

    public componentDidUpdate(): void {
        localStorage.setItem('light-theme', this.state.isLightTheme + '')
        fetchCurrentUser(this.state.isLightTheme).subscribe(undefined, error => {
            console.error(error)
            this.setState({ error })
        })
    }

    private renderBackfillRedirector = (props: RouteComponentProps<any>) => (
        <div className={'theme ' + (this.state.isLightTheme ? 'theme-light' : 'theme-dark')}>
            <BackfillRedirector {...props} onToggleTheme={this.onToggleTheme} isLightTheme={this.state.isLightTheme} />
        </div>
    )

    public render(): JSX.Element | null {
        if (this.state.error) {
            return <HeroPage icon={ErrorIcon} title={'Something happened'} subtitle={this.state.error.message} />
        }

        if (window.pageError && window.pageError.statusCode !== 404) {
            const statusText = window.pageError.statusText
            const errorMessage = window.pageError.error
            const errorID = window.pageError.errorID

            let subtitle: JSX.Element | undefined
            if (errorID) {
                subtitle = (
                    <p>
                        Sorry, there's been a problem. Please <a href="mailto:support@sourcegraph.com">contact us</a>{' '}
                        and include the error ID:
                        <span className="error-id">{errorID}</span>
                    </p>
                )
            }
            if (errorMessage) {
                subtitle = (
                    <div className="app__error">
                        {subtitle}
                        {subtitle && <hr />}
                        <pre>{errorMessage}</pre>
                    </div>
                )
            } else {
                subtitle = <div className="app__error">{subtitle}</div>
            }
            return <HeroPage icon={ServerIcon} title={'500: ' + statusText} subtitle={subtitle} />
        }

        if (window.context.licenseStatus !== 'valid' && window.location.pathname !== '/.admin/license-unverified') {
            return (
                <BrowserRouter>
                    <Switch>
                        <Route path="/.admin/license-unverified" component={BackfillRedirector} />
                        <Redirect to="/.admin/license-unverified" />
                    </Switch>
                </BrowserRouter>
            )
        }

        return (
            <BrowserRouter>
                <Route path="/" component={undefined} render={this.renderBackfillRedirector} />
            </BrowserRouter>
        )
    }

    /**
     * toggles light theme display of the container
     */
    private onToggleTheme = () => {
        this.setState(
            state => ({ isLightTheme: !state.isLightTheme }),
            () => {
                eventLogger.log(this.state.isLightTheme ? 'LightThemeClicked' : 'DarkThemeClicked')
            }
        )
    }
}

window.addEventListener('DOMContentLoaded', () => {
    render(<App />, document.querySelector('#root'))
})

updateUserSessionStores()
