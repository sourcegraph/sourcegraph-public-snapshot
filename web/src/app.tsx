// Polyfill URL because Chrome and Firefox are not spec-compliant
// Hostnames of URIs with custom schemes (e.g. git) are not parsed out
import { URL, URLSearchParams } from 'whatwg-url'
Object.assign(window, { URL, URLSearchParams })

import ErrorIcon from '@sourcegraph/icons/lib/Error'
import ServerIcon from '@sourcegraph/icons/lib/Server'
import * as React from 'react'
import { render } from 'react-dom'
import { Redirect, Route, RouteComponentProps, Switch } from 'react-router'
import { BrowserRouter } from 'react-router-dom'
import { Subscription } from 'rxjs/Subscription'
import { currentUser } from './auth'
import { HeroPage } from './components/HeroPage'
import './components/tooltip/Tooltip'
import { Tooltip } from './components/tooltip/Tooltip'
import { LinkExtension } from './extension/Link'
import { GlobalAlerts } from './global/GlobalAlerts'
import { IntegrationsToast } from './marketing/IntegrationsToast'
import { updateUserSessionStores } from './marketing/util'
import { Navbar } from './nav/Navbar'
import { routes } from './routes'
import { parseSearchURLQuery } from './search'
import { toggleSearchFilter } from './search/helpers'
import { eventLogger } from './tracking/eventLogger'

interface LayoutProps extends RouteComponentProps<any> {
    user: GQL.IUser | null
    isLightTheme: boolean
    onThemeChange: () => void
    navbarSearchQuery: string
    onNavbarQueryChange: (query: string) => void
    onFilterChosen: (value: string) => void
}

const Layout: React.SFC<LayoutProps> = props => {
    const isSearchHomepage = props.location.pathname === '/search' && !parseSearchURLQuery(props.location.search)

    const needsSiteInit = window.context.showOnboarding
    const isSiteInit = props.location.pathname === '/site-admin/init'

    const hideNavbar = isSearchHomepage || isSiteInit
    const canSyncBrowserExtension = localStorage.getItem('SYNC_BROWSER_EXT_TO_SERVER')

    return (
        <div className={`layout theme ${props.isLightTheme ? 'theme-light' : 'theme-dark'}`}>
            <GlobalAlerts isSiteAdmin={!!props.user && props.user.siteAdmin} />
            {!needsSiteInit && !isSiteInit && !!props.user && <IntegrationsToast history={props.history} />}
            {!hideNavbar && <Navbar {...props} />}
            {needsSiteInit && !isSiteInit && <Redirect to="/site-admin/init" />}
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
                            render={routeComponentProps => (
                                <div
                                    className={[
                                        'layout__app-router-container',
                                        `layout__app-router-container--${isFullWidth ? 'full-width' : 'restricted'}`,
                                    ].join(' ')}
                                >
                                    {Component && (
                                        <Component {...props} {...routeComponentProps} isFullWidth={isFullWidth} />
                                    )}
                                    {route.render && route.render({ ...props, ...routeComponentProps })}
                                    {canSyncBrowserExtension && !!props.user && <LinkExtension user={props.user} />}
                                </div>
                            )}
                        />
                    )
                })}
            </Switch>
        </div>
    )
}

interface AppState {
    error?: Error
    user?: GQL.IUser | null

    /**
     * Whether the light theme is enabled or not
     */
    isLightTheme: boolean

    /**
     * The current search query in the navbar.
     */
    navbarSearchQuery: string
}

const LIGHT_THEME_LOCAL_STORAGE_KEY = 'light-theme'

/**
 * The root component
 */
class App extends React.Component<{}, AppState> {
    public state: AppState = {
        isLightTheme: localStorage.getItem(LIGHT_THEME_LOCAL_STORAGE_KEY) !== 'false',
        navbarSearchQuery: '',
    }

    private subscriptions = new Subscription()

    public componentDidMount(): void {
        this.subscriptions.add(
            currentUser.subscribe(user => this.setState({ user }), error => this.setState({ user: null }))
        )
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public componentDidUpdate(): void {
        localStorage.setItem(LIGHT_THEME_LOCAL_STORAGE_KEY, this.state.isLightTheme + '')
    }

    public render(): React.ReactFragment | null {
        if (this.state.error) {
            return <HeroPage icon={ErrorIcon} title={'Something happened'} subtitle={this.state.error.message} />
        }

        if (window.pageError && window.pageError.statusCode !== 404) {
            const statusCode = window.pageError.statusCode
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
            return <HeroPage icon={ServerIcon} title={`${statusCode}: ${statusText}`} subtitle={subtitle} />
        }

        if (this.state.user === undefined) {
            return null
        }

        return [
            <BrowserRouter key={0}>
                <Route path="/" render={this.renderLayout} />
            </BrowserRouter>,
            <Tooltip key={1} />,
        ]
    }

    private renderLayout = (props: RouteComponentProps<any>) => (
        <Layout
            {...props}
            /* Checked for undefined in render() above */
            user={this.state.user as GQL.IUser | null}
            isLightTheme={this.state.isLightTheme}
            onThemeChange={this.onThemeChange}
            navbarSearchQuery={this.state.navbarSearchQuery}
            onNavbarQueryChange={this.onNavbarQueryChange}
            onFilterChosen={this.onFilterChosen}
        />
    )

    private onThemeChange = () => {
        this.setState(
            state => ({ isLightTheme: !state.isLightTheme }),
            () => {
                eventLogger.log(this.state.isLightTheme ? 'LightThemeClicked' : 'DarkThemeClicked')
            }
        )
    }

    private onNavbarQueryChange = (query: string) => {
        this.setState(state => ({ navbarSearchQuery: query }))
    }

    // Used for search scopes and dynamic filters
    private onFilterChosen = (searchFilter: string): void => {
        this.setState(state => ({ navbarSearchQuery: toggleSearchFilter(state.navbarSearchQuery, searchFilter) }))
    }
}

window.addEventListener('DOMContentLoaded', () => {
    render(<App />, document.querySelector('#root'))
})

updateUserSessionStores()
