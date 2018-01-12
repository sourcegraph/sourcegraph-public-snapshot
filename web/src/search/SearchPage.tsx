import * as H from 'history'
import * as React from 'react'
import { Subscription } from 'rxjs/Subscription'
import { PageTitle } from '../components/PageTitle'
import { NavLinks } from '../nav/NavLinks'
import { colorTheme, getColorTheme } from '../settings/theme'
import { eventLogger } from '../tracking/eventLogger'
import { limitString } from '../util'
import { submitSearch } from './helpers'
import { parseSearchURLQuery } from './index'
import { QueryInput } from './QueryInput'
import { SavedQueries } from './SavedQueries'
import { SearchButton } from './SearchButton'
import { SearchHelp } from './SearchHelp'
import { SearchSuggestionChips } from './SearchSuggestionChips'

interface Props {
    location: H.Location
    history: H.History
}

interface State {
    /** The query value entered by the user in the query input */
    userQuery: string

    /** The query value derived from the search fields */
    fieldsQuery: string

    isLightTheme: boolean
}

/**
 * The search page
 */
export class SearchPage extends React.Component<Props, State> {
    private subscriptions = new Subscription()

    constructor(props: Props) {
        super(props)

        const searchOptions = parseSearchURLQuery(props.location.search)
        this.state = {
            userQuery: (searchOptions && searchOptions.query) || '',
            fieldsQuery: '',
            isLightTheme: getColorTheme() === 'light',
        }
    }

    public componentDidMount(): void {
        this.subscriptions.add(colorTheme.subscribe(theme => this.setState({ isLightTheme: theme === 'light' })))
        eventLogger.logViewEvent('Home')
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        return (
            <div className="search-page">
                <PageTitle title={this.getPageTitle()} />
                <NavLinks location={this.props.location} />
                <img
                    className="search-page__logo"
                    src={
                        `${window.context.assetsRoot}/img/ui2/sourcegraph` +
                        (this.state.isLightTheme ? '-light' : '') +
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
                            global={true}
                        />
                        <SearchButton />
                        <SearchHelp />
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
                    <SavedQueries {...this.props} />
                </div>
            </div>
        )
    }

    private onSuggestionChosen = (query: string) => {
        this.setState(state => ({ userQuery: [state.userQuery, query].filter(s => s).join(' ') + ' ' }))
    }

    private onUserQueryChange = (userQuery: string) => {
        this.setState({ userQuery })
    }

    private onSubmit = (event: React.FormEvent<HTMLFormElement>): void => {
        event.preventDefault()
        const query = [this.state.fieldsQuery, this.state.userQuery].filter(s => !!s).join(' ')
        submitSearch(this.props.history, { query })
    }

    private getPageTitle(): string | undefined {
        const options = parseSearchURLQuery(this.props.location.search)
        if (options && options.query) {
            return `${limitString(this.state.userQuery, 25, true)}`
        }
        return undefined
    }
}
