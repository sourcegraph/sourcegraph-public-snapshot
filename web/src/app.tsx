
// Polyfill URL because Chrome and Firefox are not spec-compliant
// Hostnames of URIs with custom schemes (e.g. git) are not parsed out
import './util/polyfill'

import ErrorIcon from '@sourcegraph/icons/lib/Error'
import ServerIcon from '@sourcegraph/icons/lib/Server'
import * as H from 'history'
import * as React from 'react'
import { render } from 'react-dom'
import { Route, RouteComponentProps, Switch } from 'react-router'
import { BrowserRouter } from 'react-router-dom'
import { fetchCurrentUser } from './auth'
import { HeroPage } from './components/HeroPage'
import { Navbar } from './nav/Navbar'
import { routes } from './routes'
import { parseSearchURLQuery } from './search/index'
import { Search } from './search/Search'
import { handleQueryEvents } from './tracking/analyticsUtils'

interface LayoutProps {
    location: H.Location
    history: H.History
}

/**
 * Defines the layout of all pages that have a navbar
 */
class Layout extends React.Component<LayoutProps, {}> {
    public render(): JSX.Element | null {
        return (
            <div className='layout'>
                <Navbar location={this.props.location} history={this.props.history} />
                <div className='layout__app-router-container'>
                    <Switch>
                        { routes.map((route, i) => <Route key={i} {...route} />) }
                    </Switch>
                </div>
            </div>
        )
    }
}

/**
 * handles rendering Search or SearchResults components based on whether or not
 * the search query (e.g. '?q=foo') is in URL.
 */
const SearchRouter = (props: RouteComponentProps<{}>): JSX.Element | null => {
    const searchOptions = parseSearchURLQuery(props.location.search)
    if (searchOptions.query) {
        return <Layout {...props} />
    }
    return <Search {...props} />
}

interface AppState {
    error?: Error
}

/**
 * The root component
 */
class App extends React.Component<{}, AppState> {

    constructor(props: {}) {
        super(props)
        this.state = {}
        // Fetch current user data
        fetchCurrentUser().subscribe(undefined, error => {
            console.error(error)
            this.setState({ error })
        })
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
                    <Route path='/search' exact={true} component={SearchRouter} />
                    <Route component={Layout} />
                </Switch>
            </BrowserRouter>
        )
    }
}

window.addEventListener('DOMContentLoaded', () => {
    render(<App />, document.querySelector('#root'))
})

handleQueryEvents(window.location.href)
