import * as H from 'history'
import * as React from 'react'
import { Link } from 'react-router-dom'
import { PageTitle } from '../components/PageTitle'
import { UserAvatar } from '../settings/user/UserAvatar'
import { eventLogger } from '../tracking/eventLogger'
import { limitString } from '../util'
import { Help } from './Help'
import { submitSearch } from './helpers'
import { parseSearchURLQuery } from './index'
import { QueryInput } from './QueryInput'
import { ScopeLabel } from './ScopeLabel'
import { SearchButton } from './SearchButton'
import { SearchFields } from './SearchFields'
import { SearchScope } from './SearchScope'

interface Props {
    location: H.Location
    history: H.History
}

interface State {
    /** The query value entered by the user in the query input */
    userQuery: string

    /** The query value of the active search scope, or undefined if it's still loading */
    scopeQuery?: string

    /** The query value derived from the search fields */
    fieldsQuery: string
}

/**
 * The search page
 */
export class SearchPage extends React.Component<Props, State> {
    constructor(props: Props) {
        super(props)
        this.state = {
            userQuery: '',
            scopeQuery: undefined,
            fieldsQuery: '',
        }
    }

    public componentDidMount(): void {
        eventLogger.logViewEvent('Home')
    }

    public render(): JSX.Element | null {
        return (
            <div className="search-page2">
                <PageTitle title={this.getPageTitle()} />
                <div className="search-page2__nav">
                    {!window.context.user && (
                        <a href="https://about.sourcegraph.com" className="search-page2__nav-link">
                            About
                        </a>
                    )}
                    <div className="search-page2__nav-auth">
                        {// if on-prem, never show a user avatar or sign-in button
                        window.context.onPrem ? null : window.context.user ? (
                            <Link to="/settings">
                                <UserAvatar size={64} />
                            </Link>
                        ) : (
                            <Link to="/sign-in" className="btn btn-primary">
                                Sign in
                            </Link>
                        )}
                    </div>
                </div>
                <img
                    className="search-page2__logo"
                    src={`${window.context.assetsRoot}/img/ui2/sourcegraph-head-logo.svg`}
                />
                <form className="search2 search-page2__container" onSubmit={this.onSubmit}>
                    <div className="search-page2__input-container">
                        <QueryInput
                            {...this.props}
                            value={this.state.userQuery}
                            onChange={this.onUserQueryChange}
                            scopeQuery={this.state.scopeQuery}
                            prependQueryForSuggestions={this.state.fieldsQuery}
                            autoFocus={'cursor-at-end'}
                        />
                        <SearchScope
                            location={this.props.location}
                            value={this.state.scopeQuery}
                            onChange={this.onScopeQueryChange}
                        />
                        <SearchButton />
                    </div>
                    <div className="search-page2__input-sub-container">
                        <ScopeLabel scopeQuery={this.state.scopeQuery} />
                        <Help />
                    </div>
                    <SearchFields onFieldsQueryChange={this.onFieldsQueryChange} />
                </form>
            </div>
        )
    }

    private onUserQueryChange = (userQuery: string) => {
        this.setState({ userQuery })
    }

    private onScopeQueryChange = (scopeQuery: string) => {
        this.setState({ scopeQuery })
    }

    private onFieldsQueryChange = (fieldsQuery: string) => {
        this.setState({ fieldsQuery })
    }

    private onSubmit = (event: React.FormEvent<HTMLFormElement>): void => {
        event.preventDefault()
        const query = [this.state.fieldsQuery, this.state.userQuery].filter(s => !!s).join(' ')
        submitSearch(this.props.history, {
            query,
            scopeQuery: this.state.scopeQuery || '',
        })
    }

    private getPageTitle(): string | undefined {
        const query = parseSearchURLQuery(this.props.location.search)
        if (query) {
            return `${limitString(this.state.userQuery, 25, true)}`
        }
        return undefined
    }
}
