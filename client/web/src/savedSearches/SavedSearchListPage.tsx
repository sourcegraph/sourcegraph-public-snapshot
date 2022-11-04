import * as React from 'react'

import { mdiMessageTextOutline, mdiCog, mdiDelete, mdiPlus } from '@mdi/js'
import { VisuallyHidden } from '@reach/visually-hidden'
import classNames from 'classnames'
import { RouteComponentProps, useLocation } from 'react-router'
import { Subject, Subscription } from 'rxjs'
import { catchError, map, mapTo, startWith, switchMap } from 'rxjs/operators'
import { useCallbackRef } from 'use-callback-ref'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { asError, ErrorLike, isErrorLike, logger } from '@sourcegraph/common'
import { SearchPatternTypeProps } from '@sourcegraph/search'
import { buildSearchURLQuery } from '@sourcegraph/shared/src/util/url'
import { Container, PageHeader, LoadingSpinner, Button, Link, Icon, Tooltip } from '@sourcegraph/wildcard'

import { PageTitle } from '../components/PageTitle'
import { SavedSearchFields } from '../graphql-operations'
import { NamespaceProps } from '../namespaces'
import { deleteSavedSearch, fetchSavedSearches } from '../search/backend'
import { useNavbarQueryState } from '../stores'
import { eventLogger } from '../tracking/eventLogger'

import styles from './SavedSearchListPage.module.scss'

interface NodeProps extends RouteComponentProps, SearchPatternTypeProps {
    savedSearch: SavedSearchFields
    onDelete: () => void
    linkRef: React.MutableRefObject<HTMLAnchorElement | null> | null
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
    private delete = new Subject<SavedSearchFields>()

    public componentDidMount(): void {
        this.subscriptions.add(
            this.delete
                .pipe(
                    switchMap(search =>
                        deleteSavedSearch(search.id).pipe(
                            mapTo(undefined),
                            catchError(error => {
                                logger.error(error)
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
                    <Icon className={styles.rowIcon} aria-hidden={true} svgPath={mdiMessageTextOutline} />
                    <Link
                        to={
                            '/search?' +
                            buildSearchURLQuery(this.props.savedSearch.query, this.props.patternType, false)
                        }
                        ref={this.props.linkRef}
                    >
                        <div className="test-saved-search-list-page-row-title">
                            <VisuallyHidden>Run saved search: </VisuallyHidden>
                            {this.props.savedSearch.description}
                        </div>
                    </Link>
                </div>
                <div>
                    <Tooltip content="Saved search settings">
                        <Button
                            className="test-edit-saved-search-button"
                            to={`${this.props.match.path}/${this.props.savedSearch.id}`}
                            variant="secondary"
                            size="sm"
                            as={Link}
                        >
                            <Icon aria-hidden={true} svgPath={mdiCog} /> Settings
                        </Button>
                    </Tooltip>{' '}
                    <Tooltip content="Delete saved search">
                        <Button
                            aria-label="Delete"
                            className="test-delete-saved-search-button"
                            onClick={this.onDelete}
                            disabled={this.state.isDeleting}
                            variant="danger"
                            size="sm"
                        >
                            <Icon aria-hidden={true} svgPath={mdiDelete} />
                        </Button>
                    </Tooltip>
                </div>
                {this.state.isDeleting && (
                    <VisuallyHidden aria-live="polite">{`Deleted saved search: ${this.props.savedSearch.description}`}</VisuallyHidden>
                )}
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
    savedSearchesOrError?: SavedSearchFields[] | ErrorLike
}

interface Props extends RouteComponentProps, NamespaceProps {}

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
                    description="Manage notifications and alerts for specific search queries."
                    actions={
                        <Button
                            to={`${this.props.match.path}/add`}
                            className="test-add-saved-search-button"
                            variant="primary"
                            as={Link}
                        >
                            <Icon aria-hidden={true} svgPath={mdiPlus} /> Add saved search
                        </Button>
                    }
                    className="mb-3"
                >
                    <PageTitle title="Saved searches" />
                    <PageHeader.Heading as="h3" styleAs="h2">
                        <PageHeader.Breadcrumb>Saved searches</PageHeader.Breadcrumb>
                    </PageHeader.Heading>
                </PageHeader>
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
    const location = useLocation<{ description?: string }>()
    const searchPatternType = useNavbarQueryState(state => state.searchPatternType)
    const callbackReference = useCallbackRef<HTMLAnchorElement>(null, ref => ref?.focus())

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
                    <SavedSearchNode
                        key={search.id}
                        linkRef={location.state?.description === search.description ? callbackReference : null}
                        {...props}
                        patternType={searchPatternType}
                        savedSearch={search}
                    />
                ))}
            </div>
        </Container>
    )
}
