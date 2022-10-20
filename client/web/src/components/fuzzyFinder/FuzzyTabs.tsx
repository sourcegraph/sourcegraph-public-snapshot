import { Dispatch, SetStateAction, useCallback, useEffect, useMemo, useRef, useState } from 'react'

import { ApolloClient } from '@apollo/client'
import * as H from 'history'

import { KEYBOARD_SHORTCUTS } from '@sourcegraph/shared/src/keyboardShortcuts/keyboardShortcuts'
import { Settings, SettingsCascadeOrError } from '@sourcegraph/shared/src/settings/settings'
import { toPrettyBlobURL } from '@sourcegraph/shared/src/util/url'
import { useSessionStorage } from '@sourcegraph/wildcard'

import { SearchIndexing } from '../../fuzzyFinder/FuzzySearch'
import { parseBrowserRepoURL } from '../../util/url'
import { Keybindings } from '../KeyboardShortcutsHelp/KeyboardShortcutsHelp'

import { createActionsFSM, getAllFuzzyActions, FuzzyActionProps } from './FuzzyActions'
import { FuzzyFiles, loadFilesFSM } from './FuzzyFiles'
import { getFuzzyFinderFeatureFlags } from './FuzzyFinderFeatureFlag'
import { FuzzyFSM } from './FuzzyFsm'
import { FuzzyRepoRevision } from './FuzzyRepoRevision'
import { FuzzyRepos } from './FuzzyRepos'
import { FuzzySymbols } from './FuzzySymbols'

class Tab {
    constructor(
        public readonly title: string,
        public readonly isEnabled: boolean,
        public readonly shortcut?: JSX.Element,
        public readonly fsm?: FuzzyFSM
    ) {}
    public withFSM(fsm: FuzzyFSM): Tab {
        return new Tab(this.title, this.isEnabled, this.shortcut, fsm)
    }
}

const defaultTabs: Tabs = {
    all: new Tab(
        'All',
        true,
        <Keybindings uppercaseOrdered={true} keybindings={KEYBOARD_SHORTCUTS.fuzzyFinder.keybindings} />
    ),
    actions: new Tab(
        'Actions',
        true,
        <Keybindings uppercaseOrdered={true} keybindings={KEYBOARD_SHORTCUTS.fuzzyFinderActions.keybindings} />
    ),
    repos: new Tab(
        'Repos',
        true,
        <Keybindings uppercaseOrdered={true} keybindings={KEYBOARD_SHORTCUTS.fuzzyFinderRepos.keybindings} />
    ),
    symbols: new Tab(
        'Symbols',
        true,
        <Keybindings uppercaseOrdered={true} keybindings={KEYBOARD_SHORTCUTS.fuzzyFinderSymbols.keybindings} />
    ),
    files: new Tab(
        'Files',
        true,
        <Keybindings uppercaseOrdered={true} keybindings={KEYBOARD_SHORTCUTS.fuzzyFinderFiles.keybindings} />
    ),
    lines: new Tab('Lines', true),
}
const hiddenKind: Tab = new Tab('Hidden', false)

// Private helper interface to abstract over tabs. Should not be exported.
interface Tabs {
    all: Tab
    actions: Tab
    repos: Tab
    symbols: Tab
    files: Tab
    lines: Tab
}

export type FuzzyTabKey = keyof Tabs

export interface FuzzyState {
    activeTab: FuzzyTabKey
    setActiveTab: Dispatch<SetStateAction<FuzzyTabKey>>
    query: string
    setQuery: Dispatch<SetStateAction<string>>
    repoRevision: FuzzyRepoRevision
    tabs: FuzzyTabs
    onClickItem: () => void
    isGlobalFiles: boolean
    toggleGlobalFiles: () => void
    isGlobalSymbols: boolean
    toggleGlobalSymbols: () => void
}

export function fuzzyIsActive(activeTab: FuzzyTabKey, repoRevision: FuzzyRepoRevision, tab: FuzzyTabKey): boolean {
    return activeTab === 'all' || tab === activeTab
}

export function fuzzyErrors(tabs: FuzzyTabs, activeTab: FuzzyTabKey, repoRevision: FuzzyRepoRevision): string[] {
    const result: string[] = []
    for (const [key, tab] of tabs.entries()) {
        if (!fuzzyIsActive(activeTab, repoRevision, key)) {
            continue
        }
        if (!tab.fsm) {
            continue
        }
        if (tab.fsm.key === 'failed') {
            result.push(tab.fsm.errorMessage)
        }
    }
    return result
}

export class FuzzyTabs {
    constructor(public readonly underlying: Tabs) {}
    public focusTabWithIncrement(activeTab: FuzzyTabKey, increment: number): FuzzyTabKey {
        const activeIndex = this.entries().findIndex(([key]) => activeTab === key)
        const nextIndex = activeIndex + increment
        return this.focusTab(nextIndex)
    }
    public focusNamedTab(tab: FuzzyTabKey): FuzzyTabKey | undefined {
        const index = this.entries().findIndex(([key]) => key === tab)
        return index !== undefined ? this.focusTab(index) : undefined
    }
    public focusTab(index: number): FuzzyTabKey {
        const [key] = this.entries().slice(index % this.entries().length)[0]
        return key
    }
    public entries(): [FuzzyTabKey, Tab][] {
        const result: [FuzzyTabKey, Tab][] = []
        for (const key of Object.keys(this.underlying)) {
            // eslint-disable-next-line @typescript-eslint/no-unsafe-member-access, @typescript-eslint/no-explicit-any
            const value = (this.underlying as any)[key as keyof Tab] as Tab
            if (value.isEnabled) {
                result.push([key as FuzzyTabKey, value])
            }
        }
        return result
    }
    public withTabs(newTabs: Partial<Tabs>): FuzzyTabs {
        return new FuzzyTabs({ ...this.underlying, ...newTabs })
    }
    public all(): Tab[] {
        return Object.values(this.underlying).filter(tab => (tab as Tab).isEnabled)
    }
    public isOnlyFilesEnabled(): boolean {
        const [[tab], ...rest] = this.entries()
        return rest.length === 0 && tab === 'files'
    }
    public isDownloading(): boolean {
        const downloadingFSM = this.all().find(tab => tab.fsm && tab.fsm.key === 'downloading')
        return downloadingFSM !== undefined
    }
    public isAllDisabled(): boolean {
        return this.all().length === 0
    }
}

export function defaultFuzzyState(): FuzzyState {
    let query = ''
    let activeTab: FuzzyTabKey = 'all'
    return {
        query,
        onClickItem: () => {},
        isGlobalFiles: false,
        toggleGlobalFiles: () => {},
        isGlobalSymbols: false,
        toggleGlobalSymbols: () => {},
        setQuery: newQuery => {
            if (typeof newQuery === 'function') {
                query = newQuery(query)
            } else {
                query = newQuery
            }
        },
        activeTab: 'all',
        setActiveTab: newActiveTab => {
            if (typeof newActiveTab === 'function') {
                activeTab = newActiveTab(activeTab)
            } else {
                activeTab = newActiveTab
            }
        },
        repoRevision: { repositoryName: '', revision: '' },
        tabs: new FuzzyTabs(defaultTabs),
    }
}
export interface FuzzyTabsProps extends FuzzyActionProps {
    settingsCascade: SettingsCascadeOrError<Settings>
    isRepositoryRelatedPage: boolean
    location: H.Location
    client?: ApolloClient<object>
    initialQuery?: string
    isVisible: boolean
}

export function useFuzzyState(props: FuzzyTabsProps, onClickItem: () => void): FuzzyState {
    const {
        isVisible,
        location: { pathname, search, hash },
        isRepositoryRelatedPage,
        client: apolloClient,
    } = props
    let { repoName = '', commitID = '', rawRevision = '' } = useMemo(() => {
        if (pathname !== '/') {
            // TODO `parseBrowserRepoURL` should not be called on non-repoURL pages.
            return parseBrowserRepoURL(pathname + search + hash)
        }

        return { repoName: '', commitID: '', rawRevision: '' }
    }, [pathname, search, hash])
    let revision = rawRevision || commitID
    if (!isRepositoryRelatedPage) {
        repoName = ''
        revision = ''
    }
    const repoRevision: FuzzyRepoRevision = useMemo(() => ({ repositoryName: repoName, revision }), [
        repoName,
        revision,
    ])

    const {
        fuzzyFinderAll,
        fuzzyFinderActions,
        fuzzyFinderRepositories,
        fuzzyFinderSymbols,
    } = getFuzzyFinderFeatureFlags(props.settingsCascade.final)

    const [globalFilesToggleCount, setGlobalFilesToggleCount] = useState(0)
    const toggleGlobalFiles = useMemo(() => () => setGlobalFilesToggleCount(old => old + 1), [
        setGlobalFilesToggleCount,
    ])
    const isGlobalFilesRef = useRef(false)
    isGlobalFilesRef.current = globalFilesToggleCount % 2 === 1
    const [isGlobalSymbols, setGlobalSymbols] = useState(false)
    const toggleGlobalSymbols = useMemo(() => () => setGlobalSymbols(old => !old), [setGlobalSymbols])
    const isGlobalSymbolsRef = useRef(isGlobalSymbols)
    isGlobalSymbolsRef.current = isGlobalSymbols
    const localFilesRef = useRef<FuzzyFSM | null>(null)

    // NOTE: the query is cached in session storage to mimic the file pickers in
    // IntelliJ (by default) and VS Code (when "Workbench > Quick Open >
    // Preserve Input" is enabled).
    const [query, setQuery] = useSessionStorage(`fuzzy-modal.query.${repoName}`, props.initialQuery || '')
    const queryRef = useRef(query)
    queryRef.current = query
    const [activeTab, setActiveTab] = useState<FuzzyTabKey>('all')

    useEffect(() => {
        setGlobalSymbols(false)
        setGlobalFilesToggleCount(0)
    }, [isVisible, activeTab, setGlobalFilesToggleCount, setGlobalSymbols])

    const fuzzyState: Omit<FuzzyState, 'tabs'> = useMemo(
        () => ({
            onClickItem,
            query,
            setQuery,
            activeTab,
            setActiveTab,
            repoRevision,
            isGlobalFiles: globalFilesToggleCount % 2 === 1,
            isGlobalSymbols,
            toggleGlobalFiles,
            toggleGlobalSymbols,
        }),
        [
            onClickItem,
            query,
            setQuery,
            activeTab,
            repoRevision,
            globalFilesToggleCount,
            isGlobalSymbols,
            toggleGlobalFiles,
            toggleGlobalSymbols,
        ]
    )
    const fuzzyStateRef = useRef(fuzzyState)
    fuzzyStateRef.current = fuzzyState

    const [tabs, setTabs] = useState<FuzzyTabs>(
        () =>
            new FuzzyTabs({
                all: fuzzyFinderAll ? defaultTabs.all : hiddenKind,
                actions: fuzzyFinderActions
                    ? defaultTabs.actions.withFSM(createActionsFSM(getAllFuzzyActions(props)))
                    : hiddenKind,
                repos: fuzzyFinderRepositories ? defaultTabs.repos : hiddenKind,
                symbols: fuzzyFinderSymbols ? defaultTabs.symbols : hiddenKind,
                files: defaultTabs.files,
                lines: hiddenKind,
            })
    )

    const repoRevisionRef = useRef<FuzzyRepoRevision>({ repositoryName: '', revision: '' })
    repoRevisionRef.current = { repositoryName: repoName, revision }

    const tabsRef = useRef(tabs)
    tabsRef.current = tabs
    const [globalFilesNameChangeCount, setGlobalFilesNameChangeCount] = useState(0)
    const globalFilesRef = useRef<FuzzyFiles | null>(null)
    if (globalFilesRef.current === null) {
        globalFilesRef.current = new FuzzyFiles(
            apolloClient,
            () => setGlobalFilesNameChangeCount(oldCount => oldCount + 1),
            repoRevisionRef,
            isGlobalFilesRef
        )
    }

    const [repositoryNameChangeCount, setRepositoryNameChangeCount] = useState(0)
    const repositoriesRef = useRef<FuzzyRepos | null>(null)
    if (fuzzyFinderRepositories && repositoriesRef.current === null) {
        repositoriesRef.current = new FuzzyRepos(apolloClient, () =>
            setRepositoryNameChangeCount(oldCount => oldCount + 1)
        )
    }
    const hasDeletedStaleRepositories = useRef(false)
    if (isVisible && !hasDeletedStaleRepositories.current) {
        hasDeletedStaleRepositories.current = true
        repositoriesRef.current?.removeStaleResults().then(
            () => {},
            () => {}
        )
    }

    const [symbolsNameCount, setSymbolNameCount] = useState(0)
    const symbolsRef = useRef<FuzzySymbols | null>(null)
    if (fuzzyFinderSymbols && symbolsRef.current === null) {
        symbolsRef.current = new FuzzySymbols(
            apolloClient,
            () => setSymbolNameCount(oldCount => oldCount + 1),
            repoRevisionRef,
            isGlobalSymbolsRef
        )
    }

    useEffect(() => {
        for (const [key, value] of tabs.entries()) {
            if (!value.fsm) {
                continue
            }
            if (value.fsm.key === 'indexing' && !value.fsm.indexing.isIndexing()) {
                continueIndexing(value.fsm.indexing)
                    .then(next => {
                        const updatedTabs: Partial<Tabs> = {}
                        updatedTabs[key] = value.withFSM(next)
                        setTabs(tabsRef.current.withTabs(updatedTabs))
                    })
                    // eslint-disable-next-line no-console
                    .catch(error => console.error(`failed to index fuzzy tab ${key}`, error))
            }
        }
    }, [tabs])

    useEffect(() => {
        if (!isVisible) {
            return
        }
        if (!fuzzyFinderSymbols) {
            return
        }
        const isSymbolActive = fuzzyIsActive(activeTab, repoRevision, 'symbols')
        if (!isSymbolActive) {
            return
        }
        const symbols = symbolsRef.current
        if (!symbols) {
            return
        }
        setTabs(
            tabsRef.current.withTabs({
                symbols: tabsRef.current.underlying.symbols.withFSM(symbols.fuzzyFSM(query)),
            })
        )
    }, [isGlobalSymbols, isVisible, activeTab, repoRevision, query, symbolsNameCount, fuzzyFinderSymbols])

    useEffect(() => {
        if (!isVisible) {
            return
        }
        if (!fuzzyFinderRepositories) {
            return
        }
        const isRepoActive = fuzzyIsActive(activeTab, repoRevision, 'repos')
        if (!isRepoActive) {
            return
        }
        const repositories = repositoriesRef.current
        if (!repositories) {
            return
        }
        setTabs(
            tabsRef.current.withTabs({
                repos: tabsRef.current.underlying.repos.withFSM(repositories.fuzzyFSM(query)),
            })
        )
    }, [isVisible, activeTab, repoRevision, query, repositoryNameChangeCount, fuzzyFinderRepositories])

    const createURL = useCallback(
        (filename: string): string =>
            toPrettyBlobURL({
                filePath: filename,
                revision,
                repoName,
            }),
        [revision, repoName]
    )

    const setFilesFSM = useCallback(
        (fsm: FuzzyFSM) => {
            setTabs(
                tabsRef.current.withTabs({
                    files: tabsRef.current.underlying.files.withFSM(fsm),
                })
            )
        },
        [setTabs]
    )

    useEffect(() => {
        if (!isVisible) {
            return
        }
        if (!repoRevision.repositoryName) {
            return
        }
        setFilesFSM({ key: 'downloading' })
        loadFilesFSM(apolloClient, repoRevision, createURL).then(
            fsm => {
                setFilesFSM(fsm)
                localFilesRef.current = fsm
            },
            () => {}
        )
    }, [isVisible, repoRevision, apolloClient, createURL, setFilesFSM])

    useEffect(() => {
        if (!isVisible) {
            return
        }
        if (repoRevision.repositoryName) {
            return
        }
        if (!globalFilesRef.current) {
            return
        }
        setFilesFSM(globalFilesRef.current.fuzzyFSM(query))
    }, [query, globalFilesNameChangeCount, isVisible, repoRevision, setFilesFSM])

    useEffect(() => {
        if (globalFilesToggleCount === 0) {
            return
        }
        if (isGlobalFilesRef.current && globalFilesRef.current) {
            setFilesFSM(globalFilesRef.current.fuzzyFSM(queryRef.current))
        } else if (!isGlobalFilesRef.current && localFilesRef.current && repoRevisionRef.current.repositoryName) {
            setFilesFSM(localFilesRef.current)
        }
    }, [globalFilesToggleCount, setFilesFSM])

    return { ...fuzzyState, tabs }
}

async function continueIndexing(indexing: SearchIndexing): Promise<FuzzyFSM> {
    const next = await indexing.continueIndexing()
    if (next.key === 'indexing') {
        return { key: 'indexing', indexing: next }
    }
    return {
        key: 'ready',
        fuzzy: next.value,
    }
}
