import { useCallback, useRef, useEffect, type FormEvent, useState, type FC, useMemo } from 'react'

import { mdiClose, mdiArrowRight, mdiStar } from '@mdi/js'
import { VisuallyHidden } from '@reach/visually-hidden'
import classNames from 'classnames'
import { BehaviorSubject, combineLatest, of, timer } from 'rxjs'
import { catchError, debounce, map, switchMap, tap } from 'rxjs/operators'

import { asError, isErrorLike } from '@sourcegraph/common'
import type { AuthenticatedUser } from '@sourcegraph/shared/src/auth'
import type { SearchContextMinimalFields } from '@sourcegraph/shared/src/graphql-operations'
import type { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { getDefaultSearchContextSpec, type SearchContextInputProps } from '@sourcegraph/shared/src/search'
import type { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'
import type { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import {
    Badge,
    Button,
    Icon,
    ButtonLink,
    Link,
    Text,
    Tooltip,
    Combobox,
    ComboboxInput,
    ComboboxList,
    ComboboxOption,
    ComboboxOptionText,
    Alert,
    useObservable,
} from '@sourcegraph/wildcard'

import styles from './SearchContextMenu.module.scss'

export interface SearchContextMenuProps
    extends Omit<SearchContextInputProps, 'setSelectedSearchContextSpec'>,
        PlatformContextProps<'requestGraphQL'>,
        TelemetryProps,
        TelemetryV2Props {
    showSearchContextManagement: boolean
    authenticatedUser: AuthenticatedUser | null
    isSourcegraphDotCom: boolean | null
    selectSearchContextSpec: (spec: string) => void
    className?: string
    onMenuClose: (isEscapeKey?: boolean) => void
    ignoreDefaultContextDoesNotExistError?: boolean
}

interface PageInfo {
    endCursor: string | null
    hasNextPage: boolean
}

interface NextPageUpdate {
    cursor: string | undefined
    query: string
}

type LoadingState = 'LOADING' | 'LOADING_NEXT_PAGE' | 'DONE' | 'ERROR'

const SEARCH_CONTEXTS_PER_PAGE_TO_LOAD = 15

export const SearchContextMenu: FC<SearchContextMenuProps> = props => {
    const {
        authenticatedUser,
        selectedSearchContextSpec,
        selectSearchContextSpec,
        getUserSearchContextNamespaces,
        fetchSearchContexts,
        onMenuClose,
        showSearchContextManagement,
        platformContext,
        telemetryService,
        telemetryRecorder,
        isSourcegraphDotCom,
        className,
    } = props

    const [loadingState, setLoadingState] = useState<LoadingState>('DONE')
    const [searchFilter, setSearchFilter] = useState('')
    const [searchContexts, setSearchContexts] = useState<SearchContextMinimalFields[]>([])
    const [lastPageInfo, setLastPageInfo] = useState<PageInfo | null>(null)

    const infiniteScrollTrigger = useRef<HTMLDivElement | null>(null)
    const infiniteScrollList = useRef<HTMLUListElement | null>(null)

    const loadNextPageUpdates = useRef(new BehaviorSubject<NextPageUpdate>({ cursor: undefined, query: '' }))

    const loadNextPage = useCallback((): void => {
        if (loadingState === 'DONE' && (!lastPageInfo || lastPageInfo.hasNextPage)) {
            loadNextPageUpdates.current.next({
                cursor: lastPageInfo?.endCursor ?? undefined,
                query: searchFilter,
            })
        }
    }, [loadNextPageUpdates, searchFilter, lastPageInfo, loadingState])

    const onSearchFilterChanged = useCallback(
        (event: FormEvent<HTMLInputElement>) => {
            const searchFilter = event ? event.currentTarget.value : ''
            setSearchFilter(searchFilter)
            loadNextPageUpdates.current.next({ cursor: undefined, query: searchFilter })
        },
        [loadNextPageUpdates, setSearchFilter]
    )

    useEffect(() => {
        if (!infiniteScrollTrigger.current || !infiniteScrollList.current) {
            return
        }
        const intersectionObserver = new IntersectionObserver(entries => entries[0].isIntersecting && loadNextPage(), {
            root: infiniteScrollList.current,
        })
        intersectionObserver.observe(infiniteScrollTrigger.current)
        return () => intersectionObserver.disconnect()
    }, [infiniteScrollTrigger, infiniteScrollList, loadNextPage])

    useEffect(() => {
        const subscription = loadNextPageUpdates.current
            .pipe(
                tap(({ cursor }) => setLoadingState(!cursor ? 'LOADING' : 'LOADING_NEXT_PAGE')),
                // Do not debounce the initial load
                debounce(({ cursor, query }) => (!cursor && query === '' ? timer(0) : timer(300))),
                switchMap(({ cursor, query }) =>
                    combineLatest([
                        of(cursor),
                        fetchSearchContexts({
                            query,
                            platformContext,
                            first: SEARCH_CONTEXTS_PER_PAGE_TO_LOAD,
                            after: cursor,
                            namespaces: getUserSearchContextNamespaces(authenticatedUser),
                            useMinimalFields: true,
                        }),
                    ])
                ),
                tap(([, searchContextsResult]) => setLastPageInfo(searchContextsResult.pageInfo)),
                catchError(error => [asError(error)])
            )
            .subscribe(result => {
                if (!isErrorLike(result)) {
                    const [cursor, searchContextsResult] = result
                    setSearchContexts(searchContexts => {
                        // Cursor is undefined when loading the first page, so we need to replace existing search contexts
                        // E.g. when a user scrolls down to the end of the list, and starts searching
                        const initialSearchContexts = !cursor ? [] : searchContexts
                        return initialSearchContexts.concat(searchContextsResult.nodes)
                    })
                    setLoadingState('DONE')
                } else {
                    setLoadingState('ERROR')
                }
            })

        return () => subscription.unsubscribe()
    }, [
        authenticatedUser,
        loadNextPageUpdates,
        setSearchContexts,
        setLastPageInfo,
        getUserSearchContextNamespaces,
        fetchSearchContexts,
        platformContext,
    ])

    const handleContextSelect = useCallback(
        (context: string): void => {
            selectSearchContextSpec(context)
            onMenuClose(true)
            telemetryService.log('SearchContextSelected')
            telemetryRecorder.recordEvent('search.contextDropdown', 'select')
        },
        [onMenuClose, selectSearchContextSpec, telemetryService, telemetryRecorder]
    )

    const defaultContextExists = useObservable(
        useMemo(() => getDefaultSearchContextSpec({ platformContext }).pipe(map(spec => !!spec)), [platformContext])
    )

    return (
        <Combobox openOnFocus={true} className={classNames(styles.container, className)} onSelect={handleContextSelect}>
            <div className={styles.title}>
                <small>Choose search context</small>
                <Button variant="icon" aria-label="Close" className={styles.titleClose} onClick={() => onMenuClose()}>
                    <Icon aria-hidden={true} svgPath={mdiClose} />
                </Button>
            </div>
            <div className={styles.header}>
                <ComboboxInput
                    type="search"
                    variant="small"
                    placeholder="Find..."
                    autoFocus={true}
                    spellCheck={false}
                    aria-label="Find a context"
                    data-testid="search-context-menu-header-input"
                    className={styles.headerInput}
                    inputClassName={styles.headerInputElement}
                    onInput={onSearchFilterChanged}
                />
            </div>
            <ComboboxList ref={infiniteScrollList} data-testid="search-context-menu-list" className={styles.list}>
                {loadingState !== 'LOADING' && (
                    <>
                        {defaultContextExists === false && !props.ignoreDefaultContextDoesNotExistError && (
                            <Alert variant="warning" className="mx-2 mt-2">
                                Your default search context is no longer available.
                                <br />
                                <Link to="/contexts">Choose a new default context.</Link>
                            </Alert>
                        )}
                        {searchContexts.map((context, index) => (
                            <>
                                {/* Separate starred and unstarred contexts */}
                                {index > 0 &&
                                    searchContexts[index - 1].viewerHasStarred &&
                                    !context.viewerHasStarred && <div className={styles.separator} />}
                                <SearchContextMenuItem
                                    key={context.id}
                                    spec={context.spec}
                                    description={context.description}
                                    query={context.query}
                                    isDefault={context.viewerHasAsDefault}
                                    selected={context.spec === selectedSearchContextSpec}
                                    starred={context.viewerHasStarred}
                                />
                            </>
                        ))}
                    </>
                )}
                {(loadingState === 'LOADING' || loadingState === 'LOADING_NEXT_PAGE') && (
                    <div data-testid="search-context-menu-item" className={styles.item}>
                        <small>Loading search contexts...</small>
                    </div>
                )}
                {loadingState === 'ERROR' && (
                    <div data-testid="search-context-menu-item" className={classNames(styles.item, styles.itemError)}>
                        <small>Error occurred while loading search contexts</small>
                    </div>
                )}
                {loadingState === 'DONE' && searchContexts.length === 0 && (
                    <div data-testid="search-context-menu-item" className={styles.item}>
                        <small>No contexts found</small>
                    </div>
                )}

                <div ref={infiniteScrollTrigger} className={styles.infiniteScrollTrigger} />
            </ComboboxList>
            {(isSourcegraphDotCom || showSearchContextManagement) && (
                <div className={styles.footer}>
                    {isSourcegraphDotCom && (
                        <>
                            <div className="d-flex col-7 px-0">
                                <Icon
                                    className={classNames('text-merged mr-1', styles.footerIcon)}
                                    size="md"
                                    aria-hidden={true}
                                    svgPath={mdiArrowRight}
                                />
                                <Text className="mb-0">
                                    To search across your team's private repositories,{' '}
                                    <Link
                                        to="https://sourcegraph.com"
                                        onClick={() => {
                                            telemetryService.log('ClickedOnEnterpriseCTA', {
                                                location: 'ContextDropDown',
                                            })
                                            telemetryRecorder.recordEvent(
                                                'search.contextDropdown.enterpriseCTA',
                                                'clicked'
                                            )
                                        }}
                                    >
                                        get Sourcegraph Enterprise
                                    </Link>
                                    .
                                </Text>
                            </div>
                        </>
                    )}
                    {showSearchContextManagement && (
                        <>
                            <div className="flex-grow-1" />
                            <ButtonLink variant="link" to="/contexts" size="sm" className={styles.footerButton}>
                                Manage contexts
                            </ButtonLink>
                        </>
                    )}
                </div>
            )}
        </Combobox>
    )
}

interface SearchContextMenuItemProps {
    spec: string
    description: string
    query: string
    selected: boolean
    isDefault: boolean
    starred: boolean
}

export const SearchContextMenuItem: FC<SearchContextMenuItemProps> = ({
    spec,
    description,
    query,
    selected,
    isDefault,
    starred,
}) => {
    const descriptionOrQuery = description.length > 0 ? description : query

    return (
        <ComboboxOption
            value={spec}
            selected={selected}
            data-testid="search-context-menu-item"
            data-search-context-spec={spec}
            data-selected={selected || undefined}
            className={classNames(styles.item, selected && styles.itemSelected)}
        >
            <Tooltip content={spec}>
                <small data-testid="search-context-menu-item-name" className={styles.itemName}>
                    <ComboboxOptionText />
                </small>
            </Tooltip>
            {descriptionOrQuery && <VisuallyHidden>,</VisuallyHidden>}{' '}
            <Tooltip content={descriptionOrQuery}>
                <small className={styles.itemDescription}>{descriptionOrQuery}</small>
            </Tooltip>
            {isDefault && (
                <>
                    <VisuallyHidden>,</VisuallyHidden>
                    <Badge variant="secondary" className={classNames('text-uppercase ml-1', styles.itemDefault)}>
                        Default
                    </Badge>
                </>
            )}
            {starred && (
                <>
                    <VisuallyHidden>, Starred</VisuallyHidden>
                    <Icon svgPath={mdiStar} className={classNames('ml-1', styles.star)} aria-hidden={true} />
                </>
            )}
        </ComboboxOption>
    )
}
