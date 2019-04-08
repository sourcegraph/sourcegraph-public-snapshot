import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import * as H from 'history'
import AddIcon from 'mdi-react/AddIcon'
import AutoFixIcon from 'mdi-react/AutoFixIcon'
import HelpCircleOutlineIcon from 'mdi-react/HelpCircleOutlineIcon'
import * as React from 'react'
import { Redirect } from 'react-router'
import { Link } from 'react-router-dom'
import { Subject, Subscription } from 'rxjs'
import { map, startWith, switchMap } from 'rxjs/operators'
import * as GQL from '../../../../shared/src/graphql/schema'
import { SettingsCascadeProps } from '../../../../shared/src/settings/settings'
import { siteFlags } from '../../site/backend'
import { ThemeProps } from '../../theme'
import { eventLogger } from '../../tracking/eventLogger'
import { fetchSavedQueries } from '../backend'
import { ExampleSearches } from './ExampleSearches'
import { SavedQuery } from './SavedQuery'
import { SavedQueryCreateForm } from './SavedQueryCreateForm'
import { SavedQueryFields } from './SavedQueryForm'

interface Props extends SettingsCascadeProps, ThemeProps {
    authenticatedUser: GQL.IUser | null
    location: H.Location
    hideExampleSearches?: boolean
    hideTitle?: boolean
}

interface State {
    savedQueries: GQL.ISavedQuery[]

    /**
     * Whether the saved query creation form is visible.
     */
    isCreating: boolean

    loading: boolean
    error?: Error

    isViewingExamples: boolean
    exampleQuery: Partial<SavedQueryFields> | null
    disableBuiltInSearches: boolean
}

const EXAMPLE_SEARCHES_CLOSED_KEY = 'example-searches-closed'

export class SavedQueries extends React.Component<Props, State> {
    public state: State = {
        savedQueries: [],
        isCreating: false,
        loading: true,
        isViewingExamples: window.context.sourcegraphDotComMode
            ? false
            : localStorage.getItem(EXAMPLE_SEARCHES_CLOSED_KEY) !== 'true',
        exampleQuery: null,
        disableBuiltInSearches: false,
    }

    private refreshRequests = new Subject<void>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        const isHomepage = this.props.location.pathname === '/search'

        this.subscriptions.add(
            this.refreshRequests
                .pipe(
                    startWith(void 0),
                    switchMap(fetchSavedQueries),
                    map(savedQueries => ({
                        savedQueries: savedQueries
                            .filter(query => !isHomepage || query.showOnHomepage)
                            .sort((a, b) => {
                                if (a.description < b.description) {
                                    return -1
                                }
                                if (a.description === b.description && a.index < b.index) {
                                    return -1
                                }
                                return 1
                            }),
                        loading: false,
                    }))
                )
                .subscribe(newState => this.setState(newState as State), err => console.error(err))
        )

        this.subscriptions.add(
            siteFlags
                .pipe(map(({ disableBuiltInSearches }) => disableBuiltInSearches))
                .subscribe(disableBuiltInSearches => {
                    this.setState({
                        // TODO: Remove the need to check sourcegraphDotComMode by adding this to config
                        disableBuiltInSearches: window.context.sourcegraphDotComMode || disableBuiltInSearches,
                    })
                })
        )
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        if (this.state.loading) {
            return <LoadingSpinner />
        }

        const isHomepage = this.props.location.pathname === '/search'
        const isPanelOpen = this.state.isViewingExamples || this.state.isCreating

        // If not logged in, redirect to sign in.
        //
        // NOTE: This class can't use the withAuthenticatedUser wrapper because we DO NOT redirect to sign-in if
        // isHomepage.
        if (!this.props.authenticatedUser && !isHomepage) {
            const newUrl = new URL(window.location.href)
            // Return to the current page after sign up/in.
            newUrl.searchParams.set('returnTo', window.location.href)
            return <Redirect to={'/sign-in' + newUrl.search} />
        }

        return (
            <div className="saved-queries">
                {!isHomepage && !this.props.hideTitle && (
                    <div>
                        <div className="saved-queries__header">
                            <h3>{!isPanelOpen && 'Saved searches'}</h3>
                            <div className="saved-queries__actions">
                                {!this.state.disableBuiltInSearches && (
                                    <button
                                        className="btn btn-link"
                                        onClick={this.toggleExamples}
                                        disabled={this.state.isViewingExamples}
                                    >
                                        <AutoFixIcon className="icon-inline" /> Discover built-in searches
                                    </button>
                                )}

                                <button
                                    className="btn btn-link"
                                    onClick={this.toggleCreating}
                                    disabled={this.state.isCreating}
                                >
                                    <AddIcon className="icon-inline" /> Add new search
                                </button>

                                <Link
                                    to="/help/user/search/saved_searches"
                                    onClick={this.onDidClickQueryHelp}
                                    className="btn btn-link"
                                >
                                    <HelpCircleOutlineIcon className="icon-inline" /> Help
                                </Link>
                            </div>
                        </div>
                        {this.state.isCreating && (
                            <SavedQueryCreateForm
                                authenticatedUser={this.props.authenticatedUser}
                                onDidCreate={this.onDidCreateSavedQuery}
                                onDidCancel={this.toggleCreating}
                                values={this.state.exampleQuery || {}}
                                settingsCascade={this.props.settingsCascade}
                            />
                        )}
                    </div>
                )}
                <div>
                    {!this.props.hideExampleSearches && !this.state.isCreating && this.state.isViewingExamples && (
                        <ExampleSearches
                            isLightTheme={this.props.isLightTheme}
                            onClose={this.toggleExamples}
                            onExampleSelected={this.onExampleSelected}
                        />
                    )}
                    {!this.state.disableBuiltInSearches &&
                        !this.props.hideExampleSearches &&
                        !this.props.hideTitle &&
                        isPanelOpen && (
                            <div className="saved-queries__header saved-queries__space">
                                <h3>Saved searches</h3>
                            </div>
                        )}
                    {!isHomepage && this.state.savedQueries.length === 0 && (
                        <p>You don't have any saved searches yet.</p>
                    )}
                    {this.state.savedQueries.map((savedQuery, i) => (
                        <SavedQuery
                            authenticatedUser={this.props.authenticatedUser}
                            key={`${savedQuery.query}-${i}`}
                            savedQuery={savedQuery}
                            onDidDuplicate={this.onDidDuplicateSavedQuery}
                            onDidUpdate={this.onDidUpdate}
                            settingsCascade={this.props.settingsCascade}
                            isLightTheme={this.props.isLightTheme}
                        />
                    ))}
                </div>
                {this.state.savedQueries.length === 0 && this.props.authenticatedUser && isHomepage && (
                    <div className="saved-query">
                        <Link to="/search/searches">
                            <div className={`saved-query__row`}>
                                <div className="saved-query-row__add-query">
                                    <AddIcon className="icon-inline" /> Add a new search to start monitoring your code
                                </div>
                            </div>
                        </Link>
                    </div>
                )}
            </div>
        )
    }

    private toggleCreating = () => {
        eventLogger.log('SavedQueriesToggleCreating', { queries: { creating: !this.state.isCreating } })
        this.setState(state => ({ isCreating: !state.isCreating, exampleQuery: null, isViewingExamples: false }))
    }

    private toggleExamples = () => {
        eventLogger.log('SavedQueriesToggleExamples', { queries: { viewingExamples: !this.state.isViewingExamples } })

        this.setState(
            state => ({
                isViewingExamples: !state.isViewingExamples,
                exampleQuery: null,
                isCreating: false,
            }),
            () => {
                if (!this.state.isViewingExamples && localStorage.getItem(EXAMPLE_SEARCHES_CLOSED_KEY) !== 'true') {
                    localStorage.setItem(EXAMPLE_SEARCHES_CLOSED_KEY, 'true')
                }
            }
        )
    }

    private onExampleSelected = (query: Partial<SavedQueryFields>) => {
        eventLogger.log('SavedQueryExampleSelected', { queries: { example: query } })
        this.setState({ isViewingExamples: false, isCreating: true, exampleQuery: query })
    }

    private onDidCreateSavedQuery = () => {
        eventLogger.log('SavedQueryCreated')
        this.setState({ isCreating: false, exampleQuery: null }, () => this.onDidUpdate())
    }

    private onDidDuplicateSavedQuery = () => {
        eventLogger.log('SavedQueryDuplicated')
    }

    private onDidClickQueryHelp = () => {
        eventLogger.log('SavedQueriesHelpButtonClicked')
    }

    private onDidUpdate = () => this.refreshRequests.next()
}

export class SavedQueriesPage extends SavedQueries {
    public componentDidMount(): void {
        super.componentDidMount()
        eventLogger.logViewEvent('SavedQueries')
    }
}
