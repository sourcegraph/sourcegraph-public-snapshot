import React, {
    useCallback,
    useRef,
    useEffect,
    KeyboardEvent as ReactKeyboardEvent,
    FormEvent,
    useMemo,
    useState,
} from 'react'

import classNames from 'classnames'
import CloseIcon from 'mdi-react/CloseIcon'
// eslint-disable-next-line no-restricted-imports
import { DropdownItem } from 'reactstrap'
import { BehaviorSubject, combineLatest, of, timer } from 'rxjs'
import { catchError, debounce, switchMap, tap } from 'rxjs/operators'

import { asError, isErrorLike } from '@sourcegraph/common'
import { SearchContextInputProps, SearchContextMinimalFields } from '@sourcegraph/search'
import { AuthenticatedUser } from '@sourcegraph/shared/src/auth'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { ISearchContext } from '@sourcegraph/shared/src/schema'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Badge, Button, useObservable, Link, Icon } from '@sourcegraph/wildcard'

import { HighlightedSearchContextSpec } from './HighlightedSearchContextSpec'

import styles from './SearchContextMenu.module.scss'

export const SearchContextMenuItem: React.FunctionComponent<
    React.PropsWithChildren<
        {
            spec: string
            description: string
            query: string
            selected: boolean
            isDefault: boolean
            selectSearchContextSpec: (spec: string) => void
            searchFilter: string
            onKeyDown: (key: string) => void
        } & TelemetryProps
    >
> = ({
    spec,
    description,
    query,
    selected,
    isDefault,
    selectSearchContextSpec,
    searchFilter,
    onKeyDown,
    telemetryService,
}) => {
    const setContext = useCallback(() => {
        telemetryService.log('SearchContextSelected')
        selectSearchContextSpec(spec)
    }, [selectSearchContextSpec, spec, telemetryService])

    const descriptionOrQuery = description.length > 0 ? description : query

    return (
        <DropdownItem
            data-testid="search-context-menu-item"
            className={classNames(styles.item, selected && styles.itemSelected)}
            onClick={setContext}
            role="menuitem"
            data-search-context-spec={spec}
            onKeyDown={event => onKeyDown(event.key)}
        >
            <small
                data-testid="search-context-menu-item-name"
                className={classNames('font-weight-medium', styles.itemName)}
                title={spec}
            >
                <HighlightedSearchContextSpec spec={spec} searchFilter={searchFilter} />
            </small>{' '}
            <small className={styles.itemDescription} title={descriptionOrQuery}>
                {descriptionOrQuery}
            </small>
            {isDefault && (
                <Badge variant="secondary" className={classNames('text-uppercase', styles.itemDefault)}>
                    Default
                </Badge>
            )}
        </DropdownItem>
    )
}

export interface SearchContextMenuProps
    extends Omit<
            SearchContextInputProps,
            'setSelectedSearchContextSpec' | 'hasUserAddedRepositories' | 'hasUserAddedExternalServices'
        >,
        PlatformContextProps<'requestGraphQL'>,
        TelemetryProps {
    showSearchContextManagement: boolean
    authenticatedUser: AuthenticatedUser | null
    closeMenu: (isEscapeKey?: boolean) => void
    selectSearchContextSpec: (spec: string) => void
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

const searchContextsPerPageToLoad = 15

const getSearchContextMenuItem = (spec: string): HTMLButtonElement | null =>
    document.querySelector(`[data-search-context-spec="${spec}"]`)

export const SearchContextMenu: React.FunctionComponent<React.PropsWithChildren<SearchContextMenuProps>> = ({
    authenticatedUser,
    selectedSearchContextSpec,
    defaultSearchContextSpec,
    selectSearchContextSpec,
    getUserSearchContextNamespaces,
    fetchAutoDefinedSearchContexts,
    fetchSearchContexts,
    closeMenu,
    showSearchContextManagement,
    platformContext,
    telemetryService,
}) => {
    const inputElement = useRef<HTMLInputElement | null>(null)

    const reset = useCallback(() => {
        selectSearchContextSpec(defaultSearchContextSpec)
        closeMenu()
    }, [closeMenu, defaultSearchContextSpec, selectSearchContextSpec])

    const focusInputElement = (): void => {
        // Focus the input in the next run-loop to override the browser focusing the first dropdown item
        // if the user opened the dropdown using a keyboard
        setTimeout(() => inputElement.current?.focus(), 0)
    }

    // Reactstrap is preventing default behavior on all non-DropdownItem elements inside a Dropdown,
    // so we need to stop propagation to allow normal behavior (e.g. enter and space to activate buttons)
    // See Reactstrap bug: https://github.com/reactstrap/reactstrap/issues/2099
    const onResetButtonKeyDown = useCallback((event: ReactKeyboardEvent<HTMLButtonElement>): void => {
        if (event.key === ' ' || event.key === 'Enter') {
            event.stopPropagation()
        }
    }, [])

    const onMenuKeyDown = useCallback(
        (event: ReactKeyboardEvent<HTMLDivElement>): void => {
            if (event.key === 'Escape') {
                closeMenu(true)
                event.stopPropagation()
            }
        },
        [closeMenu]
    )

    const [loadingState, setLoadingState] = useState<LoadingState>('DONE')
    const [searchFilter, setSearchFilter] = useState('')
    const [searchContexts, setSearchContexts] = useState<SearchContextMinimalFields[]>([])
    const [lastPageInfo, setLastPageInfo] = useState<PageInfo | null>(null)

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
        const subscription = loadNextPageUpdates.current
            .pipe(
                tap(({ cursor }) => setLoadingState(!cursor ? 'LOADING' : 'LOADING_NEXT_PAGE')),
                // Do not debounce the initial load
                debounce(({ cursor, query }) => (!cursor && query === '' ? timer(0) : timer(300))),
                switchMap(({ cursor, query }) =>
                    combineLatest([
                        of(cursor),
                        fetchSearchContexts({
                            first: searchContextsPerPageToLoad,
                            query,
                            after: cursor,
                            namespaces: getUserSearchContextNamespaces(authenticatedUser),
                            platformContext,
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
    const filteredList = useMemo(() => filteredAutoDefinedSearchContexts.concat(searchContexts as ISearchContext[]), [
        filteredAutoDefinedSearchContexts,
        searchContexts,
    ])

    const infiniteScrollTrigger = useRef<HTMLDivElement | null>(null)
    const infiniteScrollList = useRef<HTMLDivElement | null>(null)
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

    useEffect(focusInputElement, [])

    const onInputKeyDown = useCallback(
        (event: React.KeyboardEvent) => {
            if (filteredList.length > 0 && event.key === 'ArrowDown') {
                getSearchContextMenuItem(filteredList[0].spec)?.focus()
                event.stopPropagation()
                event.preventDefault()
            }
        },
        [filteredList]
    )

    return (
        // eslint-disable-next-line jsx-a11y/no-static-element-interactions
        <div onKeyDown={onMenuKeyDown}>
            <div className={styles.title}>
                <small>Choose search context</small>
                <Button onClick={() => closeMenu()} variant="icon" className={styles.titleClose} aria-label="Close">
                    <Icon as={CloseIcon} />
                </Button>
            </div>
            <div className={classNames('d-flex', styles.header)}>
                <input
                    ref={inputElement}
                    onInput={onSearchFilterChanged}
                    onKeyDown={onInputKeyDown}
                    type="search"
                    placeholder="Find..."
                    aria-label="Find a context"
                    data-testid="search-context-menu-header-input"
                    className={classNames('form-control form-control-sm', styles.headerInput)}
                />
            </div>
            <div data-testid="search-context-menu-list" className={styles.list} ref={infiniteScrollList} role="menu">
                {loadingState !== 'LOADING' &&
                    filteredList.map((context, index) => (
                        <SearchContextMenuItem
                            key={context.id}
                            spec={context.spec}
                            description={context.description}
                            query={context.query}
                            isDefault={context.spec === defaultSearchContextSpec}
                            selected={context.spec === selectedSearchContextSpec}
                            selectSearchContextSpec={selectSearchContextSpec}
                            searchFilter={searchFilter}
                            onKeyDown={key => index === 0 && key === 'ArrowUp' && focusInputElement()}
                            telemetryService={telemetryService}
                        />
                    ))}
                {(loadingState === 'LOADING' || loadingState === 'LOADING_NEXT_PAGE') && (
                    <DropdownItem data-testid="search-context-menu-item" className={styles.item} disabled={true}>
                        <small>Loading search contexts...</small>
                    </DropdownItem>
                )}
                {loadingState === 'ERROR' && (
                    <DropdownItem
                        data-testid="search-context-menu-item"
                        className={classNames(styles.item, styles.itemError)}
                        disabled={true}
                    >
                        <small>Error occured while loading search contexts</small>
                    </DropdownItem>
                )}
                {loadingState === 'DONE' && filteredList.length === 0 && (
                    <DropdownItem data-testid="search-context-menu-item" className={styles.item} disabled={true}>
                        <small>No contexts found</small>
                    </DropdownItem>
                )}
                {/* Dummy element to prevent a focus error when using the keyboard to open the dropdown */}
                <DropdownItem className="d-none" />
                <div ref={infiniteScrollTrigger} className={styles.infiniteScrollTrigger} />
            </div>
            <div className={styles.footer}>
                <Button
                    onClick={reset}
                    onKeyDown={onResetButtonKeyDown}
                    className={styles.footerButton}
                    variant="link"
                    size="sm"
                >
                    Reset
                </Button>
                <span className="flex-grow-1" />
                {showSearchContextManagement && (
                    <Button
                        to="/contexts"
                        className={styles.footerButton}
                        onClick={() => closeMenu()}
                        variant="link"
                        size="sm"
                        as={Link}
                    >
                        Manage contexts
                    </Button>
                )}
            </div>
        </div>
    )
}
