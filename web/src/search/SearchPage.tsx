import * as H from 'history'
import * as React from 'react'
import { PageTitle } from '../components/PageTitle'
import { NavLinks } from '../nav/NavLinks'
import { eventLogger } from '../tracking/eventLogger'
import { limitString } from '../util'
import { queryIndexOfScope, submitSearch } from './helpers'
import { parseSearchURLQuery } from './index'
import { QueryInput } from './QueryInput'
import { SavedQueries } from './SavedQueries'
import { SearchButton } from './SearchButton'
import { SearchHelp } from './SearchHelp'
import { SearchSuggestionChips } from './SearchSuggestionChips'

interface Props {
    user: GQL.IUser | null
    location: H.Location
    history: H.History
    isLightTheme: boolean
    onThemeChange: () => void
}

interface State {
    /** The query value entered by the user in the query input */
    userQuery: string

    /** The query value derived from the search fields */
    fieldsQuery: string
}

/**
 * The search page
 */
export class SearchPage extends React.Component<Props, State> {
    private static HIDE_REPOGROUP_SAMPLE_STORAGE_KEY = 'SearchPage/hideRepogroupSample'

    constructor(props: Props) {
        super(props)

        const searchOptions = parseSearchURLQuery(props.location.search)
        this.state = {
            userQuery: (searchOptions && searchOptions.query) || '',
            fieldsQuery: '',
        }
    }

    public componentDidMount(): void {
        eventLogger.logViewEvent('Home')
        if (
            window.context.sourcegraphDotComMode &&
            !localStorage.getItem(SearchPage.HIDE_REPOGROUP_SAMPLE_STORAGE_KEY)
        ) {
            this.addScopeToQuery('repogroup:sample')
        }
    }

    public render(): JSX.Element | null {
        return (
            <div className="search-page">
                <PageTitle title={this.getPageTitle()} />
                <NavLinks {...this.props} />
                <img
                    className="search-page__logo"
                    src={
                        `${window.context.assetsRoot}/img/sourcegraph` +
                        (this.props.isLightTheme ? '-light' : '') +
                        '-head-logo.svg'
                    }
                />
                <form className="search search-page__container" onSubmit={this.onSubmit}>
                    <div className="search-page__input-container">
                        <QueryInput
                            {...this.props}
                            value={this.state.userQuery}
                            onChange={this.onUserQueryChange}
                            prependQueryForSuggestions={this.state.fieldsQuery}
                            autoFocus={'cursor-at-end'}
                            hasGlobalQueryBehavior={true}
                        />
                        <SearchButton />
                        <SearchHelp className="search-page__help" />
                    </div>
                    <div className="search-page__input-sub-container">
                        <SearchSuggestionChips
                            location={this.props.location}
                            onSuggestionChosen={this.onSuggestionChosen}
                            query={this.state.userQuery}
                        />
                    </div>
                </form>
                <div className="search search-page__query-container">
                    <SavedQueries {...this.props} hideExampleSearches={true} />
                </div>
            </div>
        )
    }

    private onSuggestionChosen = (scope: string) => {
        const idx = queryIndexOfScope(this.state.userQuery, scope)
        if (idx === -1) {
            this.addScopeToQuery(scope)
        } else {
            this.removeScopeFromQuery(scope, idx)
        }
    }

    private addScopeToQuery(scope: string): void {
        this.setState(state => ({ userQuery: [state.userQuery.trim(), scope].filter(s => s).join(' ') + ' ' }))
        if (window.context.sourcegraphDotComMode && scope === 'repogroup:sample') {
            localStorage.removeItem(SearchPage.HIDE_REPOGROUP_SAMPLE_STORAGE_KEY)
        }
    }

    private removeScopeFromQuery(scope: string, idx: number): void {
        this.setState(state => ({
            userQuery: (
                state.userQuery.substring(0, idx).trim() +
                ' ' +
                state.userQuery.substring(idx + scope.length).trim()
            ).trim(),
        }))

        if (window.context.sourcegraphDotComMode && scope === 'repogroup:sample') {
            localStorage.setItem(SearchPage.HIDE_REPOGROUP_SAMPLE_STORAGE_KEY, 'true')
        }
    }

    private onUserQueryChange = (userQuery: string) => {
        this.setState({ userQuery })

        if (window.context.sourcegraphDotComMode) {
            if (queryIndexOfScope(userQuery, 'repogroup:sample') !== -1) {
                localStorage.removeItem(SearchPage.HIDE_REPOGROUP_SAMPLE_STORAGE_KEY)
            } else {
                localStorage.setItem(SearchPage.HIDE_REPOGROUP_SAMPLE_STORAGE_KEY, 'true')
            }
        }
    }

    private onSubmit = (event: React.FormEvent<HTMLFormElement>): void => {
        event.preventDefault()
        const query = [this.state.fieldsQuery, this.state.userQuery].filter(s => !!s).join(' ')
        submitSearch(this.props.history, { query }, 'home')
    }

    private getPageTitle(): string | undefined {
        const options = parseSearchURLQuery(this.props.location.search)
        if (options && options.query) {
            return `${limitString(this.state.userQuery, 25, true)}`
        }
        return undefined
    }
}
