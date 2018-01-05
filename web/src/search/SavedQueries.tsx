import AddIcon from '@sourcegraph/icons/lib/Add'
import HelpIcon from '@sourcegraph/icons/lib/Help'
import Loader from '@sourcegraph/icons/lib/Loader'
import * as H from 'history'
import * as React from 'react'
import { Link } from 'react-router-dom'
import { map } from 'rxjs/operators/map'
import { Subject } from 'rxjs/Subject'
import { Subscription } from 'rxjs/Subscription'
import { colorTheme, getColorTheme } from '../settings/theme'
import { eventLogger } from '../tracking/eventLogger'
import { observeSavedQueries } from './backend'
import { SavedQuery } from './SavedQuery'
import { SavedQueryCreateForm } from './SavedQueryCreateForm'

interface Props {
    location: H.Location
}

interface State {
    savedQueries: GQL.ISavedQuery[]

    /**
     * Whether the saved query creation form is visible.
     */
    creating: boolean

    loading: boolean
    error?: Error
    isLightTheme: boolean
}

export class SavedQueries extends React.Component<Props, State> {
    public state: State = {
        savedQueries: [],
        creating: false,
        loading: true,
        isLightTheme: getColorTheme() === 'light',
    }

    private componentUpdates = new Subject<Props>()
    private subscriptions = new Subscription()

    constructor(props: Props) {
        super(props)

        this.subscriptions.add(
            observeSavedQueries()
                .pipe(
                    map(savedQueries => ({
                        savedQueries: savedQueries.sort((a, b) => {
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
    }

    public componentDidMount(): void {
        this.subscriptions.add(colorTheme.subscribe(theme => this.setState({ isLightTheme: theme === 'light' })))
        eventLogger.logViewEvent('SavedQueries')
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

        const savedQueries = this.state.savedQueries.filter(savedQuery => {
            if (isHomepage) {
                return savedQuery.viewOnHomepage
            }
            return savedQuery
        })

        return (
            <div className="saved-queries">
                {!isHomepage && (
                    <div>
                        <div className="saved-queries__header">
                            <h2>Saved queries</h2>
                            <span className="saved-queries__center">
                                <button
                                    className="btn btn-link"
                                    onClick={this.toggleCreating}
                                    disabled={this.state.creating}
                                >
                                    <AddIcon className="icon-inline" /> Add new query
                                </button>

                                <a
                                    className="saved-queries__help"
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
                        {this.state.creating && (
                            <SavedQueryCreateForm
                                onDidCreate={this.onDidCreateSavedQuery}
                                onDidCancel={this.toggleCreating}
                            />
                        )}
                        {!this.state.creating &&
                            this.state.savedQueries.length === 0 && <p>You don't have any saved queries yet.</p>}
                    </div>
                )}
                <div>
                    {savedQueries.map((savedQuery, i) => (
                        <SavedQuery
                            hideBottomBorder={i === 0 && savedQueries.length > 1}
                            key={i}
                            savedQuery={savedQuery}
                            onDidDuplicate={this.onDidDuplicateSavedQuery}
                        />
                    ))}
                </div>
                {savedQueries.length === 0 &&
                    isHomepage && (
                        <div className="saved-query">
                            <Link to="/search/queries">
                                <div className={`saved-query__row`}>
                                    <div className="saved-query__add-query">
                                        <AddIcon className="icon-inline" /> Add a new query to start monitoring your
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
        eventLogger.log('SavedQueriesToggleCreating', { queries: { creating: !this.state.creating } })
        this.setState({ creating: !this.state.creating })
    }

    private onDidCreateSavedQuery = () => {
        eventLogger.log('SavedQueryCreated')
        this.setState({ creating: false })
    }

    private onDidDuplicateSavedQuery = () => {
        eventLogger.log('SavedQueryDuplicated')
    }
}
