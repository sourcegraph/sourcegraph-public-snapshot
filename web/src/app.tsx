// Polyfill URL because Chrome and Firefox are not spec-compliant
// Hostnames of URIs with custom schemes (e.g. git) are not parsed out
import { URL, URLSearchParams } from 'whatwg-url'
// The polyfill does not expose createObjectURL, which we need for creating data: URIs for Web
// Workers. So retain it.
//
// tslint:disable-next-line:no-unbound-method
const createObjectURL = window.URL ? window.URL.createObjectURL : null
Object.assign(window, { URL, URLSearchParams })
;(window.URL.createObjectURL as any) = createObjectURL

// Load only a subset of the highlight.js languages
import { registerLanguage } from 'highlight.js/lib/highlight'
registerLanguage('go', require('highlight.js/lib/languages/go'))
registerLanguage('javascript', require('highlight.js/lib/languages/javascript'))
registerLanguage('typescript', require('highlight.js/lib/languages/typescript'))
registerLanguage('java', require('highlight.js/lib/languages/java'))
registerLanguage('python', require('highlight.js/lib/languages/python'))
registerLanguage('php', require('highlight.js/lib/languages/php'))
registerLanguage('bash', require('highlight.js/lib/languages/bash'))
registerLanguage('clojure', require('highlight.js/lib/languages/clojure'))
registerLanguage('cpp', require('highlight.js/lib/languages/cpp'))
registerLanguage('cs', require('highlight.js/lib/languages/cs'))
registerLanguage('css', require('highlight.js/lib/languages/css'))
registerLanguage('dockerfile', require('highlight.js/lib/languages/dockerfile'))
registerLanguage('elixir', require('highlight.js/lib/languages/elixir'))
registerLanguage('haskell', require('highlight.js/lib/languages/haskell'))
registerLanguage('html', require('highlight.js/lib/languages/xml'))
registerLanguage('lua', require('highlight.js/lib/languages/lua'))
registerLanguage('ocaml', require('highlight.js/lib/languages/ocaml'))
registerLanguage('r', require('highlight.js/lib/languages/r'))
registerLanguage('ruby', require('highlight.js/lib/languages/ruby'))
registerLanguage('rust', require('highlight.js/lib/languages/rust'))
registerLanguage('swift', require('highlight.js/lib/languages/swift'))

import ErrorIcon from '@sourcegraph/icons/lib/Error'
import ServerIcon from '@sourcegraph/icons/lib/Server'
import { Component as CXPComponent, EMPTY_ENVIRONMENT as CXP_EMPTY_ENVIRONMENT } from 'cxp/lib/environment/environment'
import { URI } from 'cxp/lib/types/textDocument'
import { isEqual } from 'lodash'
import * as React from 'react'
import { render } from 'react-dom'
import { Redirect, Route, RouteComponentProps, Switch } from 'react-router'
import { BrowserRouter } from 'react-router-dom'
import { Subscription } from 'rxjs'
import { currentUser } from './auth'
import { Extensions, ExtensionsChangeProps, ExtensionsProps } from './backend/features'
import * as GQL from './backend/graphqlschema'
import { FeedbackText } from './components/FeedbackText'
import { HeroPage } from './components/HeroPage'
import { Tooltip } from './components/tooltip/Tooltip'
import { CONTROLLER } from './cxp/controller'
import { CXPEnvironmentProps, CXPProps } from './cxp/CXPEnvironment'
import { LinkExtension } from './extension/Link'
import { GlobalAlerts } from './global/GlobalAlerts'
import { IntegrationsToast } from './marketing/IntegrationsToast'
import { updateUserSessionStores } from './marketing/util'
import { GlobalNavbar } from './nav/GlobalNavbar'
import { routes } from './routes'
import { parseSearchURLQuery } from './search'
import { eventLogger } from './tracking/eventLogger'

interface LayoutProps extends RouteComponentProps<any>, ExtensionsProps, ExtensionsChangeProps, CXPProps {
    user: GQL.IUser | null
    isLightTheme: boolean
    onThemeChange: () => void
    navbarSearchQuery: string
    onNavbarQueryChange: (query: string) => void
    showHelpPopover: boolean
    onHelpPopoverToggle: (visible?: boolean) => void
}

const Layout: React.SFC<LayoutProps> = props => {
    const isSearchHomepage = props.location.pathname === '/search' && !parseSearchURLQuery(props.location.search)

    const needsSiteInit = window.context.showOnboarding
    const isSiteInit = props.location.pathname === '/site-admin/init'

    // Force light theme on site init page.
    if (isSiteInit && !props.isLightTheme) {
        props.onThemeChange()
    }

    // Remove trailing slash (which is never valid in any of our URLs).
    if (props.location.pathname !== '/' && props.location.pathname.endsWith('/')) {
        return <Redirect to={{ ...props.location, pathname: props.location.pathname.slice(0, -1) }} />
    }

    return (
        <div className="layout">
            <GlobalAlerts isSiteAdmin={!!props.user && props.user.siteAdmin} />
            {!needsSiteInit && !isSiteInit && !!props.user && <IntegrationsToast history={props.history} />}
            {!isSiteInit && <GlobalNavbar {...props} lowProfile={isSearchHomepage} />}
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
                                    {!!props.user && <LinkExtension user={props.user} />}
                                </div>
                            )}
                        />
                    )
                })}
            </Switch>
        </div>
    )
}

interface AppState extends ExtensionsProps, CXPEnvironmentProps {
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

    /** Whether the help popover is shown. */
    showHelpPopover: boolean
}

const LIGHT_THEME_LOCAL_STORAGE_KEY = 'light-theme'

/**
 * The root component
 */
class App extends React.Component<{}, AppState> {
    public state: AppState = {
        isLightTheme: localStorage.getItem(LIGHT_THEME_LOCAL_STORAGE_KEY) !== 'false',
        navbarSearchQuery: '',
        showHelpPopover: false,
        extensions: [],
        cxpEnvironment: CXP_EMPTY_ENVIRONMENT,
    }

    private subscriptions = new Subscription()

    public componentDidMount(): void {
        document.body.classList.add('theme')
        this.subscriptions.add(
            currentUser.subscribe(user => this.setState({ user }), error => this.setState({ user: null }))
        )
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
        document.body.classList.remove('theme')
        document.body.classList.remove('theme-light')
        document.body.classList.remove('theme-dark')
    }

    public componentDidUpdate(): void {
        localStorage.setItem(LIGHT_THEME_LOCAL_STORAGE_KEY, this.state.isLightTheme + '')
        document.body.classList.toggle('theme-light', this.state.isLightTheme)
        document.body.classList.toggle('theme-dark', !this.state.isLightTheme)
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
                subtitle = <FeedbackText headerText="Sorry, there's been a problem." />
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
            showHelpPopover={this.state.showHelpPopover}
            onHelpPopoverToggle={this.onHelpPopoverToggle}
            extensions={this.state.extensions}
            onExtensionsChange={this.onExtensionsChange}
            cxpEnvironment={this.state.cxpEnvironment}
            cxpOnComponentChange={this.cxpOnComponentChange}
            cxpOnRootChange={this.cxpOnRootChange}
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

    private onNavbarQueryChange = (navbarSearchQuery: string) => {
        this.setState({ navbarSearchQuery })
    }

    private onHelpPopoverToggle = (visible?: boolean): void => {
        eventLogger.log('HelpPopoverToggled')
        this.setState(prevState => ({
            // If visible is any non-boolean type (e.g., MouseEvent), treat it as undefined. This lets callers use
            // onHelpPopoverToggle directly in an event handler without wrapping it in an another function.
            showHelpPopover: visible !== true && visible !== false ? !prevState.showHelpPopover : visible,
        }))
    }

    private onExtensionsChange = (extensions: Extensions): void => {
        this.setState(
            prevState =>
                isEqual(prevState.extensions, extensions)
                    ? null
                    : {
                          extensions,
                          cxpEnvironment: {
                              ...prevState.cxpEnvironment,
                              extensions: extensions.map(x => ({
                                  id: x.extensionID,
                                  settings: x.settings,
                                  manifest: x.manifest,
                              })),
                          },
                      },
            () => CONTROLLER.setEnvironment(this.state.cxpEnvironment)
        )
    }

    private cxpOnComponentChange = (component: CXPComponent | null): void => {
        this.setState(
            prevState => ({ cxpEnvironment: { ...prevState.cxpEnvironment, component } }),
            () => CONTROLLER.setEnvironment(this.state.cxpEnvironment)
        )
    }

    private cxpOnRootChange = (root: URI | null): void => {
        this.setState(
            prevState => ({ cxpEnvironment: { ...prevState.cxpEnvironment, root } }),
            () => CONTROLLER.setEnvironment(this.state.cxpEnvironment)
        )
    }
}

window.addEventListener('DOMContentLoaded', () => {
    render(<App />, document.querySelector('#root'))
})

updateUserSessionStores()
