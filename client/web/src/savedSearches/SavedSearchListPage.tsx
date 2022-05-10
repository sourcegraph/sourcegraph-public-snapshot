import * as React from 'react'

import classNames from 'classnames'
import DeleteIcon from 'mdi-react/DeleteIcon'
import MessageTextOutlineIcon from 'mdi-react/MessageTextOutlineIcon'
import PlusIcon from 'mdi-react/PlusIcon'
import SettingsIcon from 'mdi-react/SettingsIcon'
import { RouteComponentProps } from 'react-router'
import { Subject, Subscription } from 'rxjs'
import { catchError, map, mapTo, startWith, switchMap } from 'rxjs/operators'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { asError, ErrorLike, isErrorLike } from '@sourcegraph/common'
import { SearchPatternTypeProps } from '@sourcegraph/search'
import * as GQL from '@sourcegraph/shared/src/schema'
import { buildSearchURLQuery } from '@sourcegraph/shared/src/util/url'
import { Container, PageHeader, LoadingSpinner, Button, Link, Icon } from '@sourcegraph/wildcard'

import { NamespaceProps } from '../namespaces'
import { deleteSavedSearch, fetchSavedSearches } from '../search/backend'
import { useNavbarQueryState } from '../stores'
import { eventLogger } from '../tracking/eventLogger'

import styles from './SavedSearchListPage.module.scss'

interface NodeProps extends RouteComponentProps, SearchPatternTypeProps {
    savedSearch: GQL.ISavedSearch
    onDelete: () => void
}

interface NodeState {
    isDeleting: boolean
}

class SavedSearchNode extends React.PureComponent<NodeProps, NodeState> {
    constructor(props: NodeProps) {
        super(props)
        this.state = { isDeleting: false }
    }

    private subscriptions = new Subscription()
    private delete = new Subject<GQL.ISavedSearch>()

    public componentDidMount(): void {
        this.subscriptions.add(
            this.delete
                .pipe(
                    switchMap(search =>
                        deleteSavedSearch(search.id).pipe(
                            mapTo(undefined),
                            catchError(error => {
                                console.error(error)
                                return []
                            })
                        )
                    )
                )
                .subscribe(() => {
                    eventLogger.log('SavedSearchDeleted')
                    this.setState({ isDeleting: false })
                    this.props.onDelete()
                })
        )
    }
    public render(): JSX.Element | null {
        return (
            <div className={classNames(styles.row, 'list-group-item test-saved-search-list-page-row')}>
                <div className="d-flex">
                    <Icon className={styles.rowIcon} as={MessageTextOutlineIcon} />
                    <Link
                        to={
                            '/search?' +
                            buildSearchURLQuery(this.props.savedSearch.query, this.props.patternType, false)
                        }
                    >
                        <div className="test-saved-search-list-page-row-title">
                            {this.props.savedSearch.description}
                        </div>
                    </Link>
                </div>
                <div>
                    <Button
                        className="test-edit-saved-search-button"
                        to={`${this.props.match.path}/${this.props.savedSearch.id}`}
                        data-tooltip="Saved search settings"
                        variant="secondary"
                        size="sm"
                        as={Link}
                    >
                        <Icon as={SettingsIcon} /> Settings
                    </Button>{' '}
                    <Button
                        className="test-delete-saved-search-button"
                        onClick={this.onDelete}
                        disabled={this.state.isDeleting}
                        data-tooltip="Delete saved search"
                        variant="danger"
                        size="sm"
                        aria-label="Delete saved search"
                    >
                        <Icon as={DeleteIcon} />
                    </Button>
                </div>
            </div>
        )
    }

    private onDelete = (): void => {
        if (!window.confirm(`Delete the saved search ${this.props.savedSearch.description}?`)) {
            return
        }
        this.setState({ isDeleting: true })
        this.delete.next(this.props.savedSearch)
    }
}

interface State {
    savedSearchesOrError?: GQL.ISavedSearch[] | ErrorLike
}

interface Props extends RouteComponentProps<{}>, NamespaceProps {}

export class SavedSearchListPage extends React.Component<Props, State> {
    public subscriptions = new Subscription()
    private refreshRequests = new Subject<void>()

    public state: State = {}

    public componentDidMount(): void {
        this.subscriptions.add(
            this.refreshRequests
                .pipe(
                    startWith(undefined),
                    switchMap(() => fetchSavedSearches().pipe(catchError(error => [asError(error)]))),
                    map(savedSearchesOrError => ({ savedSearchesOrError }))
                )
                .subscribe(newState => this.setState(newState as State))
        )
        eventLogger.logViewEvent('SavedSearchListPage')
    }

    public render(): JSX.Element | null {
        return (
            <div className={styles.savedSearchListPage} data-testid="saved-searches-list-page">
                <PageHeader
                    path={[{ text: 'Saved searches' }]}
                    headingElement="h2"
                    description="Manage notifications and alerts for specific search queries."
                    actions={
                        <Button
                            to={`${this.props.match.path}/add`}
                            className="test-add-saved-search-button"
                            variant="primary"
                            as={Link}
                        >
                            <Icon as={PlusIcon} /> Add saved search
                        </Button>
                    }
                    className="mb-3"
                />
                <SavedSearchListPageContent onDelete={this.onDelete} {...this.props} {...this.state} />
            </div>
        )
    }

    private onDelete = (): void => {
        this.refreshRequests.next()
    }
}

interface SavedSearchListPageContentProps extends Props, State {
    onDelete: () => void
}

const SavedSearchListPageContent: React.FunctionComponent<React.PropsWithChildren<SavedSearchListPageContentProps>> = ({
    namespace,
    savedSearchesOrError,
    ...props
}) => {
    const searchPatternType = useNavbarQueryState(state => state.searchPatternType)

    if (savedSearchesOrError === undefined) {
        return <LoadingSpinner />
    }

    if (isErrorLike(savedSearchesOrError)) {
        return <ErrorAlert className="mb-3" error={savedSearchesOrError} />
    }

    const namespaceSavedSearches = savedSearchesOrError.filter(search => namespace.id === search.namespace.id)
    if (namespaceSavedSearches.length === 0) {
        return <Container className="text-center text-muted">You haven't created a saved search yet.</Container>
    }

    return (
        <Container>
            <div className="list-group list-group-flush">
                {namespaceSavedSearches.map(search => (
                    <SavedSearchNode key={search.id} {...props} patternType={searchPatternType} savedSearch={search} />
                ))}
            </div>
        </Container>
    )
}
