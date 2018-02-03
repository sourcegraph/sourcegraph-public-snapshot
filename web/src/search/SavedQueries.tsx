import AddIcon from '@sourcegraph/icons/lib/Add'
import HelpIcon from '@sourcegraph/icons/lib/Help'
import Loader from '@sourcegraph/icons/lib/Loader'
import WandIcon from '@sourcegraph/icons/lib/MagicWand'
import * as H from 'history'
import * as React from 'react'
import { Redirect } from 'react-router'
import { Link } from 'react-router-dom'
import { map } from 'rxjs/operators/map'
import { Subject } from 'rxjs/Subject'
import { Subscription } from 'rxjs/Subscription'
import { currentUser } from '../auth'
import { eventLogger } from '../tracking/eventLogger'
import { observeSavedQueries } from './backend'
import { ExampleSearches } from './ExampleSearches'
import { SavedQuery } from './SavedQuery'
import { SavedQueryCreateForm } from './SavedQueryCreateForm'
import { SavedQueryFields } from './SavedQueryForm'

interface Props {
    location: H.Location
    isLightTheme: boolean
    hideExampleSearches: boolean
}

interface State {
    savedQueries: GQL.ISavedQuery[]

    /**
     * Whether the saved query creation form is visible.
     */
    isCreating: boolean

    loading: boolean
    error?: Error
    user: GQL.IUser | null

    isViewingExamples: boolean
    exampleQuery: Partial<SavedQueryFields> | null
}

const EXAMPLE_SEARCHES_CLOSED_KEY = 'example-searches-closed'

export class SavedQueries extends React.Component<Props, State> {
    public state: State = {
        savedQueries: [],
        isCreating: false,
        loading: true,
        user: null,
        isViewingExamples: window.context.sourcegraphDotComMode
            ? false
            : localStorage.getItem(EXAMPLE_SEARCHES_CLOSED_KEY) !== 'true',
        exampleQuery: null,
    }

    private componentUpdates = new Subject<Props>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        const isHomepage = this.props.location.pathname === '/search'

        this.subscriptions.add(
            observeSavedQueries()
                .pipe(
                    map(savedQueries => ({
                        savedQueries: savedQueries.filter(query => !isHomepage || query.showOnHomepage).sort((a, b) => {
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

        this.subscriptions.add(currentUser.subscribe(user => this.setState({ user })))
    }

    public componentWillReceiveProps(newProps: Props): void {
        this.componentUpdates.next(newProps)
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        if (this.state.loading) {
            return <Loader />
        }

        const isHomepage = this.props.location.pathname === '/search'
        const isPanelOpen = this.state.isViewingExamples || this.state.isCreating

        // If not logged in, redirect to sign in
        if (!this.state.user && !isHomepage) {
            const newUrl = new URL(window.location.href)
            // Return to the current page after sign up/in.
            newUrl.searchParams.set('returnTo', window.location.href)
            return <Redirect to={'/sign-up' + newUrl.search} />
        }

        return (
            <div className="saved-queries">
                {!isHomepage && (
                    <div>
                        <div className="saved-queries__header">
                            <h2>{!isPanelOpen && 'Saved searches'}</h2>
                            <span className="saved-queries__center">
                                {!window.context.sourcegraphDotComMode && (
                                    <button
                                        className="btn btn-link saved-queries__btn"
                                        onClick={this.toggleExamples}
                                        disabled={this.state.isViewingExamples}
                                    >
                                        <WandIcon className="icon-inline saved-queries__wand" />
                                        Discover built-in searches
                                    </button>
                                )}

                                <button
                                    className="btn btn-link saved-queries__btn"
                                    onClick={this.toggleCreating}
                                    disabled={this.state.isCreating}
                                >
                                    <AddIcon className="icon-inline" /> Add new search
                                </button>

                                <a
                                    onClick={this.onDidClickQueryHelp}
                                    className="saved-queries__help saved-queries__btn"
                                    href="https://about.sourcegraph.com/docs/search/#saved-queries"
                                    target="_blank"
                                >
                                    <small>
                                        <HelpIcon className="icon-inline" />
                                        <span>Help</span>
                                    </small>
                                </a>
                            </span>
                        </div>
                        {this.state.isCreating && (
                            <SavedQueryCreateForm
                                onDidCreate={this.onDidCreateSavedQuery}
                                onDidCancel={this.toggleCreating}
                                values={this.state.exampleQuery || {}}
                            />
                        )}
                        {!this.state.isCreating &&
                            !this.state.isViewingExamples &&
                            this.state.savedQueries.length === 0 && <p>You don't have any saved searches yet.</p>}
                    </div>
                )}
                <div>
                    {!this.props.hideExampleSearches &&
                        !this.state.isCreating &&
                        this.state.isViewingExamples && (
                            <ExampleSearches
                                isLightTheme={this.props.isLightTheme}
                                onClose={this.toggleExamples}
                                onExampleSelected={this.onExampleSelected}
                            />
                        )}
                    {!this.props.hideExampleSearches &&
                        isPanelOpen && (
                            <div className="saved-queries__header saved-queries__space">
                                <h2>Saved searches</h2>
                            </div>
                        )}
                    {this.state.savedQueries.map((savedQuery, i) => (
                        <SavedQuery
                            key={`${savedQuery.query.query}-${i}`}
                            savedQuery={savedQuery}
                            onDidDuplicate={this.onDidDuplicateSavedQuery}
                            isLightTheme={this.props.isLightTheme}
                        />
                    ))}
                </div>
                {this.state.savedQueries.length === 0 &&
                    this.state.user &&
                    isHomepage && (
                        <div className="saved-query">
                            <Link to="/search/searches">
                                <div className={`saved-query__row`}>
                                    <div className="saved-query__add-query">
                                        <AddIcon className="icon-inline" /> Add a new search to start monitoring your
                                        code.
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
        this.setState({ isCreating: false, exampleQuery: null })
    }

    private onDidDuplicateSavedQuery = () => {
        eventLogger.log('SavedQueryDuplicated')
    }

    private onDidClickQueryHelp = () => {
        eventLogger.log('SavedQueriesHelpButtonClicked')
    }
}

export class SavedQueriesPage extends SavedQueries {
    public componentDidMount(): void {
        super.componentDidMount()
        eventLogger.logViewEvent('SavedQueries')
    }
}
