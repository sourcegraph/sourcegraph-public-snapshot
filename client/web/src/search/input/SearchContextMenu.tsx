import classNames from 'classnames'
import CloseIcon from 'mdi-react/CloseIcon'
import React, {
    useCallback,
    useRef,
    useEffect,
    KeyboardEvent as ReactKeyboardEvent,
    FormEvent,
    useMemo,
    useState,
} from 'react'
import { DropdownItem } from 'reactstrap'
import { BehaviorSubject, combineLatest, of, timer } from 'rxjs'
import { catchError, debounce, switchMap, tap } from 'rxjs/operators'

import { Link } from '@sourcegraph/shared/src/components/Link'
import { ISearchContext } from '@sourcegraph/shared/src/graphql/schema'
import { asError, isErrorLike } from '@sourcegraph/shared/src/util/errors'
import { useObservable } from '@sourcegraph/shared/src/util/useObservable'

import { SearchContextInputProps } from '..'
import { AuthenticatedUser } from '../../auth'
import { SearchContextFields } from '../../graphql-operations'
import { eventLogger } from '../../tracking/eventLogger'

import { HighlightedSearchContextSpec } from './HighlightedSearchContextSpec'
import styles from './SearchContextMenu.module.scss'

export const SearchContextMenuItem: React.FunctionComponent<{
    spec: string
    description: string
    selected: boolean
    isDefault: boolean
    selectSearchContextSpec: (spec: string) => void
    searchFilter: string
}> = ({ spec, description, selected, isDefault, selectSearchContextSpec, searchFilter }) => {
    const setContext = useCallback(() => {
        eventLogger.log('SearchContextSelected')
        selectSearchContextSpec(spec)
    }, [selectSearchContextSpec, spec])
    return (
        <DropdownItem
            data-testid="search-context-menu-item"
            className={classNames(styles.item, selected && styles.itemSelected)}
            onClick={setContext}
        >
            <small
                data-testid="search-context-menu-item-name"
                className={classNames('font-weight-medium', styles.itemName)}
                title={spec}
            >
                <HighlightedSearchContextSpec spec={spec} searchFilter={searchFilter} />
            </small>{' '}
            <small className={styles.itemDescription} title={description}>
                {description}
            </small>
            {isDefault && (
                <span className={classNames('badge badge-secondary text-uppercase', styles.itemDefault)}>Default</span>
            )}
        </DropdownItem>
    )
}

export interface SearchContextMenuProps
    extends Omit<
        SearchContextInputProps,
        | 'showSearchContext'
        | 'setSelectedSearchContextSpec'
        | 'hasUserAddedRepositories'
        | 'hasUserAddedExternalServices'
    > {
    authenticatedUser: AuthenticatedUser | null
    closeMenu: () => void
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

const getFirstMenuItem = (): HTMLButtonElement | null => document.querySelector(`${styles.item}:first-child`)

export const SearchContextMenu: React.FunctionComponent<SearchContextMenuProps> = ({
    authenticatedUser,
    selectedSearchContextSpec,
    defaultSearchContextSpec,
    selectSearchContextSpec,
    getUserSearchContextNamespaces,
    fetchAutoDefinedSearchContexts,
    fetchSearchContexts,
    closeMenu,
    showSearchContextManagement,
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

    useEffect(() => {
        focusInputElement()
        const onInputKeyDown = (event: KeyboardEvent): void => {
            if (event.key === 'ArrowDown') {
                getFirstMenuItem()?.focus()
                event.stopPropagation()
            }
        }
        const currentInput = inputElement.current
        currentInput?.addEventListener('keydown', onInputKeyDown)
        return () => currentInput?.removeEventListener('keydown', onInputKeyDown)
    }, [])

    useEffect(() => {
        const firstMenuItem = getFirstMenuItem()
        const onFirstMenuItemKeyDown = (event: KeyboardEvent): void => {
            if (event.key === 'ArrowUp') {
                focusInputElement()
                event.stopPropagation()
            }
        }
        firstMenuItem?.addEventListener('keydown', onFirstMenuItemKeyDown)
        return () => firstMenuItem?.removeEventListener('keydown', onFirstMenuItemKeyDown)
    }, [])

    const onMenuKeyDown = useCallback(
        (event: ReactKeyboardEvent<HTMLDivElement>): void => {
            if (event.key === 'Escape') {
                closeMenu()
                event.stopPropagation()
            }
        },
        [closeMenu]
    )

    const [loadingState, setLoadingState] = useState<LoadingState>('DONE')
    const [searchFilter, setSearchFilter] = useState('')
    const [searchContexts, setSearchContexts] = useState<SearchContextFields[]>([])
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
    ])

    const autoDefinedSearchContexts = useObservable(
        useMemo(() => fetchAutoDefinedSearchContexts().pipe(catchError(error => [asError(error)])), [
            fetchAutoDefinedSearchContexts,
        ])
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

    return (
        // eslint-disable-next-line jsx-a11y/no-static-element-interactions
        <div onKeyDown={onMenuKeyDown}>
            <div className={styles.title}>
                <small>Choose search context</small>
                <button
                    onClick={closeMenu}
                    type="button"
                    className={classNames('btn btn-icon', styles.titleClose)}
                    aria-label="Close"
                >
                    <CloseIcon className="icon-inline" />
                </button>
            </div>
            <div className={classNames('d-flex', styles.header)}>
                <input
                    ref={inputElement}
                    onInput={onSearchFilterChanged}
                    type="search"
                    placeholder="Find..."
                    aria-label="Find a context"
                    data-testid="search-context-menu-header-input"
                    className={classNames('form-control form-control-sm', styles.headerInput)}
                />
            </div>
            <div data-testid="search-context-menu-list" className={styles.list} ref={infiniteScrollList} role="menu">
                {loadingState !== 'LOADING' &&
                    filteredList.map(context => (
                        <SearchContextMenuItem
                            key={context.id}
                            spec={context.spec}
                            description={context.description}
                            isDefault={context.spec === defaultSearchContextSpec}
                            selected={context.spec === selectedSearchContextSpec}
                            selectSearchContextSpec={selectSearchContextSpec}
                            searchFilter={searchFilter}
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
                <div ref={infiniteScrollTrigger} className={styles.infiniteScrollTrigger} />
            </div>
            <div className={styles.footer}>
                <button
                    type="button"
                    onClick={reset}
                    onKeyDown={onResetButtonKeyDown}
                    className={classNames('btn btn-link btn-sm', styles.footerButton)}
                >
                    Reset
                </button>
                <span className="flex-grow-1" />
                {showSearchContextManagement && (
                    <Link
                        to="/contexts"
                        className={classNames('btn btn-link btn-sm', styles.footerButton)}
                        onClick={closeMenu}
                    >
                        Manage contexts
                    </Link>
                )}
            </div>
        </div>
    )
}
