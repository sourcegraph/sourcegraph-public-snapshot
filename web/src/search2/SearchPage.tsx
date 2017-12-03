import * as H from 'history'
import * as React from 'react'
import { PageTitle } from '../components/PageTitle'
import { NavLinks } from '../nav/NavLinks'
import { eventLogger } from '../tracking/eventLogger'
import { limitString } from '../util'
import { submitSearch } from './helpers'
import { parseSearchURLQuery } from './index'
import { QueryInput } from './QueryInput'
import { ScopeLabel } from './ScopeLabel'
import { Search2Help } from './Search2Help'
import { SearchButton } from './SearchButton'
import { SearchFields } from './SearchFields'
import { SearchScope } from './SearchScope'

interface Props {
    location: H.Location
    history: H.History
    onToggleTheme: () => void
    isLightTheme: boolean
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

        const searchOptions = parseSearchURLQuery(props.location.search)
        this.state = {
            userQuery: '',
            scopeQuery: searchOptions ? searchOptions.scopeQuery : undefined,
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
                <NavLinks onToggleTheme={this.props.onToggleTheme} isLightTheme={this.props.isLightTheme} />
                <img
                    className="search-page2__logo"
                    src={
                        `${window.context.assetsRoot}/img/ui2/sourcegraph` +
                        (this.props.isLightTheme ? '-light' : '') +
                        '-head-logo.svg'
                    }
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
                        <Search2Help />
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
            scopeQuery: this.state.scopeQuery,
        })
    }

    private getPageTitle(): string | undefined {
        const options = parseSearchURLQuery(this.props.location.search)
        if (options && options.query) {
            return `${limitString(this.state.userQuery, 25, true)}`
        }
        return undefined
    }
}
