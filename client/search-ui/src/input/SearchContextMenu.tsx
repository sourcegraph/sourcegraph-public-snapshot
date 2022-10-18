import { useCallback, useRef, useEffect, FormEvent, useMemo, useState, FC } from 'react'

import { mdiClose } from '@mdi/js'
import classNames from 'classnames'
import { BehaviorSubject, combineLatest, of, timer } from 'rxjs'
import { catchError, debounce, switchMap, tap } from 'rxjs/operators'

import { asError, isErrorLike } from '@sourcegraph/common'
import { SearchContextInputProps, SearchContextMinimalFields } from '@sourcegraph/search'
import { AuthenticatedUser } from '@sourcegraph/shared/src/auth'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import {
    Badge,
    Button,
    useObservable,
    Icon,
    ButtonLink,
    Tooltip,
    Combobox,
    ComboboxInput,
    ComboboxList,
    ComboboxOption,
    ComboboxOptionText,
} from '@sourcegraph/wildcard'

import styles from './SearchContextMenu.module.scss'

export interface SearchContextMenuProps
    extends Omit<SearchContextInputProps, 'setSelectedSearchContextSpec'>,
        PlatformContextProps<'requestGraphQL'>,
        TelemetryProps {
    showSearchContextManagement: boolean
    authenticatedUser: AuthenticatedUser | null
    selectSearchContextSpec: (spec: string) => void
    className?: string
    onMenuClose: (isEscapeKey?: boolean) => void
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
        defaultSearchContextSpec,
        selectSearchContextSpec,
        getUserSearchContextNamespaces,
        fetchAutoDefinedSearchContexts,
        fetchSearchContexts,
        onMenuClose,
        showSearchContextManagement,
        platformContext,
        telemetryService,
        className,
    } = props

    const [loadingState, setLoadingState] = useState<LoadingState>('DONE')
    const [searchFilter, setSearchFilter] = useState('')
    const [searchContexts, setSearchContexts] = useState<SearchContextMinimalFields[]>([])
    const [lastPageInfo, setLastPageInfo] = useState<PageInfo | null>(null)

    const infiniteScrollTrigger = useRef<HTMLDivElement | null>(null)
    const infiniteScrollList = useRef<HTMLUListElement | null>(null)

    const loadNextPageUpdates = useRef(
        new BehaviorSubject<NextPageUpdate>({ cursor: undefined, query: '' })
    )

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

    const autoDefinedSearchContexts = useObservable(
        useMemo(
            () =>
                fetchAutoDefinedSearchContexts({ platformContext, useMinimalFields: true }).pipe(
                    catchError(error => [asError(error)])
                ),
            [fetchAutoDefinedSearchContexts, platformContext]
        )
    )

    const reset = useCallback(() => {
        selectSearchContextSpec(defaultSearchContextSpec)
        onMenuClose()
    }, [onMenuClose, defaultSearchContextSpec, selectSearchContextSpec])

    const handleContextSelect = useCallback(
        (context: string): void => {
            selectSearchContextSpec(context)
            onMenuClose(true)
            telemetryService.log('SearchContextSelected')
        },
        [onMenuClose, selectSearchContextSpec, telemetryService]
    )

    const filteredAutoDefinedSearchContexts = useMemo(
        () =>
            autoDefinedSearchContexts && !isErrorLike(autoDefinedSearchContexts)
                ? autoDefinedSearchContexts.filter(context =>
                      context.spec.toLowerCase().includes(searchFilter.toLowerCase())
                  )
                : [],
        [autoDefinedSearchContexts, searchFilter]
    )

    // Merge auto-defined contexts and user-defined contexts
    const filteredList = useMemo(() => filteredAutoDefinedSearchContexts.concat(searchContexts), [
        filteredAutoDefinedSearchContexts,
        searchContexts,
    ])

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
                {loadingState !== 'LOADING' &&
                    filteredList.map(context => (
                        <SearchContextMenuItem
                            key={context.id}
                            spec={context.spec}
                            description={context.description}
                            query={context.query}
                            isDefault={context.spec === defaultSearchContextSpec}
                            selected={context.spec === selectedSearchContextSpec}
                        />
                    ))}
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
                {loadingState === 'DONE' && filteredList.length === 0 && (
                    <div data-testid="search-context-menu-item" className={styles.item}>
                        <small>No contexts found</small>
                    </div>
                )}

                <div ref={infiniteScrollTrigger} className={styles.infiniteScrollTrigger} />
            </ComboboxList>
            <div className={styles.footer}>
                <Button size="sm" variant="link" className={styles.footerButton} onClick={reset}>
                    Reset
                </Button>
                <span className="flex-grow-1" />
                {showSearchContextManagement && (
                    <ButtonLink variant="link" to="/contexts" size="sm" className={styles.footerButton}>
                        Manage contexts
                    </ButtonLink>
                )}
            </div>
        </Combobox>
    )
}

interface SearchContextMenuItemProps {
    spec: string
    description: string
    query: string
    selected: boolean
    isDefault: boolean
}

export const SearchContextMenuItem: FC<SearchContextMenuItemProps> = props => {
    const { spec, description, query, selected, isDefault } = props

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
            </Tooltip>{' '}
            <Tooltip content={descriptionOrQuery}>
                <small className={styles.itemDescription}>{descriptionOrQuery}</small>
            </Tooltip>
            {isDefault && (
                <Badge variant="secondary" className={classNames('text-uppercase', styles.itemDefault)}>
                    Default
                </Badge>
            )}
        </ComboboxOption>
    )
}
