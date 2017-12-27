// Polyfill URL because Chrome and Firefox are not spec-compliant
// Hostnames of URIs with custom schemes (e.g. git) are not parsed out
import './util/polyfill'

import ErrorIcon from '@sourcegraph/icons/lib/Error'
import ServerIcon from '@sourcegraph/icons/lib/Server'
import * as React from 'react'
import { render } from 'react-dom'
import { Redirect, Route, RouteComponentProps, Switch } from 'react-router'
import { BrowserRouter } from 'react-router-dom'
import { Subscription } from 'rxjs/Subscription'
import { currentUser } from './auth'
import { HeroPage } from './components/HeroPage'
import { updateUserSessionStores } from './marketing/util'
import { Navbar } from './nav/Navbar'
import { routes } from './routes'
import { parseSearchURLQuery } from './search'
import { colorTheme, getColorTheme } from './settings/theme'

interface LayoutProps extends RouteComponentProps<any> {
    isLightTheme: boolean
    user: GQL.IUser | null
}

const Layout: React.SFC<LayoutProps> = props => {
    const isSearchHomepage = props.location.pathname === '/search' && !parseSearchURLQuery(props.location.search)

    const needsSiteInit = window.context.onPrem && window.context.showOnboarding
    const isSiteInit = props.location.pathname === '/site-admin/init'

    const hideNavbar = isSearchHomepage || isSiteInit

    const transferProps = { user: props.user }

    return (
        <div className={`layout theme ${props.isLightTheme ? 'theme-light' : 'theme-dark'}`}>
            {!hideNavbar && <Navbar location={props.location} history={props.history} />}
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
                            render={props => (
                                <div
                                    className={[
                                        'layout__app-router-container',
                                        `layout__app-router-container--${isFullWidth ? 'full-width' : 'restricted'}`,
                                    ].join(' ')}
                                >
                                    {Component && <Component {...props} {...transferProps} isFullWidth={isFullWidth} />}
                                    {route.render && route.render({ ...props, ...transferProps })}
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
    isLightTheme: boolean
}

/**
 * The root component
 */
class App extends React.Component<{}, AppState> {
    public state: AppState = {
        isLightTheme: getColorTheme() === 'light',
    }

    private subscriptions = new Subscription()

    public componentDidMount(): void {
        this.subscriptions.add(currentUser.subscribe(user => this.setState({ user })))
        this.subscriptions.add(colorTheme.subscribe(theme => this.setState({ isLightTheme: theme === 'light' })))
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

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

        if (this.state.user === undefined) {
            return null
        }

        return (
            <BrowserRouter>
                <Route path="/" render={this.renderLayout} />
            </BrowserRouter>
        )
    }

    private renderLayout = (props: LayoutProps) => (
        <Layout {...props} user={this.state.user!} isLightTheme={this.state.isLightTheme} />
    )
}

window.addEventListener('DOMContentLoaded', () => {
    render(<App />, document.querySelector('#root'))
})

updateUserSessionStores()
