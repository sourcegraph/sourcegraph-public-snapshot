import * as React from 'react'

import { mdiMessageTextOutline, mdiCog, mdiDelete, mdiPlus } from '@mdi/js'
import { VisuallyHidden } from '@reach/visually-hidden'
import classNames from 'classnames'
import { RouteComponentProps, useLocation } from 'react-router'
import { Subject, Subscription } from 'rxjs'
import { catchError, mapTo, switchMap } from 'rxjs/operators'
import { useCallbackRef } from 'use-callback-ref'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { ErrorLike, isErrorLike, logger } from '@sourcegraph/common'
import { SearchPatternTypeProps } from '@sourcegraph/search'
import { buildSearchURLQuery } from '@sourcegraph/shared/src/util/url'
import { Container, PageHeader, LoadingSpinner, Button, Link, Icon, Tooltip } from '@sourcegraph/wildcard'

import {
    usePaginatedConnection,
    PaginationProps,
    PaginatedConnection,
} from '../components/FilteredConnection/hooks/usePaginatedConnection'
import { PageTitle } from '../components/PageTitle'
import { SavedSearchFields, SavedSearchesPageResult, SavedSearchesPageVariables } from '../graphql-operations'
import { NamespaceProps } from '../namespaces'
import { deleteSavedSearch, SAVED_SEARCHES_PAGE_QUERY } from '../search/backend'
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

const PAGE_SIZE = 5

interface Props extends RouteComponentProps, NamespaceProps {}

export function SavedSearchListPage(props: Props) {
    const [currentPage, setCurrentPage] = React.useState(1)
    React.useEffect(() => {
        eventLogger.logViewEvent('SavedSearchListPage')
    }, [])

    const { connection, error, ...pagination } = usePaginatedConnection<
        SavedSearchesPageResult,
        SavedSearchesPageVariables,
        SavedSearchFields
    >({
        query: SAVED_SEARCHES_PAGE_QUERY,
        options: { pageSize: PAGE_SIZE },
        variables: {
            namespaceType: props.namespace.__typename,
            namespaceId: props.namespace.id,
        },
        getConnection: ({ data }) => console.log(data) || data?.savedSearches,
    })

    console.log(pagination)

    return (
        <div className={styles.savedSearchListPage} data-testid="saved-searches-list-page">
            <PageHeader
                description="Manage notifications and alerts for specific search queries."
                actions={
                    <Button
                        to={`${props.match.path}/add`}
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
            <SavedSearchListPageContent
                onDelete={() => {
                    //refetch()
                }}
                pagination={pagination}
                currentPage={currentPage}
                savedSearches={connection}
                error={error}
                {...props}
            />
        </div>
    )
}

interface SavedSearchListPageContentProps extends Props {
    error?: ErrorLike
    savedSearches?: PaginatedConnection<SavedSearchFields>
    onDelete: () => void
    currentPage: number
    pagination: PaginationProps
}

const SavedSearchListPageContent: React.FunctionComponent<React.PropsWithChildren<SavedSearchListPageContentProps>> = ({
    error,
    savedSearches,
    currentPage,
    pagination,
    ...props
}) => {
    const location = useLocation<{ description?: string }>()
    const searchPatternType = useNavbarQueryState(state => state.searchPatternType)
    const callbackReference = useCallbackRef<HTMLAnchorElement>(null, ref => ref?.focus())

    if (isErrorLike(error)) {
        return <ErrorAlert className="mb-3" error={error} />
    }

    if (savedSearches === undefined) {
        return <LoadingSpinner />
    }

    if (savedSearches.totalCount === 0) {
        return <Container className="text-center text-muted">You haven't created a saved search yet.</Container>
    }

    return (
        <Container>
            <div className="list-group list-group-flush mb-4">
                {savedSearches.nodes.map(search => (
                    <SavedSearchNode
                        key={search.id}
                        linkRef={location.state?.description === search.description ? callbackReference : null}
                        {...props}
                        patternType={searchPatternType}
                        savedSearch={search}
                    />
                ))}
            </div>
            {/*
            <PageSelector
                currentPage={currentPage}
                totalPages={Math.ceil(savedSearches.totalCount / PAGE_SIZE)}
                onPageChange={onPageChange}
            />
          */}
            <button type="button" onClick={pagination.goToFirstPage}>
                First page
            </button>
            {savedSearches?.pageInfo?.hasNextPage && (
                <button type="button" onClick={pagination.goToNextPage}>
                    Next page
                </button>
            )}
            {savedSearches?.pageInfo?.hasPreviousPage && (
                <button type="button" onClick={pagination.goToPreviousPage}>
                    Previous page
                </button>
            )}
            <button type="button" onClick={pagination.goToLastPage}>
                Last page
            </button>
        </Container>
    )
}
