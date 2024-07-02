import { useCallback, useEffect, useMemo, useRef, useState, type Dispatch, type SetStateAction } from 'react'

import type { ApolloClient } from '@apollo/client'
import type * as H from 'history'

import { KEYBOARD_SHORTCUTS } from '@sourcegraph/shared/src/keyboardShortcuts/keyboardShortcuts'
import type { Settings, SettingsCascadeOrError } from '@sourcegraph/shared/src/settings/settings'
import { useTheme } from '@sourcegraph/shared/src/theme'
import { toPrettyBlobURL } from '@sourcegraph/shared/src/util/url'
import { useSessionStorage } from '@sourcegraph/wildcard'

import { SearchValueRankingCache } from '../../fuzzyFinder/SearchValueRankingCache'
import { parseBrowserRepoURL } from '../../util/url'
import { Keybindings, plaintextKeybindings } from '../KeyboardShortcutsHelp/KeyboardShortcutsHelp'
import type { UserHistory } from '../useUserHistory'

import { createActionsFSM, getAllFuzzyActions } from './FuzzyActions'
import { FuzzyFiles, FuzzyRepoFiles } from './FuzzyFiles'
import { useFuzzyFinderFeatureFlags } from './FuzzyFinderFeatureFlag'
import type { FuzzyFSM } from './FuzzyFsm'
import type { FuzzyRepoRevision } from './FuzzyRepoRevision'
import { FuzzyRepos } from './FuzzyRepos'
import { FuzzySymbols } from './FuzzySymbols'

class Tab {
    constructor(
        public readonly title: string,
        public readonly isEnabled: boolean,
        public readonly shortcut?: JSX.Element,
        public readonly plaintextShortcut?: string
    ) {}
}

const defaultTabs: Tabs = {
    all: new Tab(
        'All',
        true,
        <Keybindings uppercaseOrdered={true} keybindings={KEYBOARD_SHORTCUTS.fuzzyFinder.keybindings} />,
        plaintextKeybindings(KEYBOARD_SHORTCUTS.fuzzyFinder.keybindings)
    ),
    actions: new Tab(
        'Actions',
        true,
        <Keybindings uppercaseOrdered={true} keybindings={KEYBOARD_SHORTCUTS.fuzzyFinderActions.keybindings} />,
        plaintextKeybindings(KEYBOARD_SHORTCUTS.fuzzyFinderActions.keybindings)
    ),
    repos: new Tab(
        'Repos',
        true,
        <Keybindings uppercaseOrdered={true} keybindings={KEYBOARD_SHORTCUTS.fuzzyFinderRepos.keybindings} />,
        plaintextKeybindings(KEYBOARD_SHORTCUTS.fuzzyFinderRepos.keybindings)
    ),
    symbols: new Tab(
        'Symbols',
        true,
        <Keybindings uppercaseOrdered={true} keybindings={KEYBOARD_SHORTCUTS.fuzzyFinderSymbols.keybindings} />,
        plaintextKeybindings(KEYBOARD_SHORTCUTS.fuzzyFinderSymbols.keybindings)
    ),
    files: new Tab(
        'Files',
        true,
        <Keybindings uppercaseOrdered={true} keybindings={KEYBOARD_SHORTCUTS.fuzzyFinderFiles.keybindings} />,
        plaintextKeybindings(KEYBOARD_SHORTCUTS.fuzzyFinderFiles.keybindings)
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

export type FuzzyScope = 'everywhere' | 'repository'

export class FuzzyTabFSM {
    constructor(
        public readonly key: FuzzyTabKey,
        public readonly scope: FuzzyScope | 'always',
        public readonly fsm: () => FuzzyFSM,
        public readonly onQuery?: (query: string) => void
    ) {}
    public isActive(activeTab: FuzzyTabKey, scope: FuzzyScope): boolean {
        const isScopeMatch = this.scope === 'always' || this.scope === scope
        const isTabMatch = activeTab === 'all' || this.key === activeTab
        return isScopeMatch && isTabMatch
    }
}

export interface FuzzyState {
    activeTab: FuzzyTabKey
    setActiveTab: Dispatch<SetStateAction<FuzzyTabKey>>
    query: string
    setQuery: (newQuery: string) => void
    repoRevision: FuzzyRepoRevision
    tabs: FuzzyTabs
    rankingCache: SearchValueRankingCache
    /**
     * fsmGeneration increases whenever `FuzzyTabs.fsms` have new underlying data
     * meaning the query should be re-triggered.
     */
    fsmGeneration: number
    setScope: Dispatch<SetStateAction<FuzzyScope>>
    isScopeToggleDisabled: boolean
    scope: FuzzyScope
    toggleScope: () => void
}

export function fuzzyIsActive(activeTab: FuzzyTabKey, tab: FuzzyTabKey): boolean {
    return activeTab === 'all' || tab === activeTab
}

export function fuzzyErrors(tabs: FuzzyTabs, activeTab: FuzzyTabKey, scope: FuzzyScope): string[] {
    const result: string[] = []
    for (const tab of tabs.fsms) {
        if (!tab.isActive(activeTab, scope)) {
            continue
        }
        const fsm = tab.fsm()
        if (fsm.key === 'failed') {
            result.push(fsm.errorMessage)
        }
    }
    return result
}

export class FuzzyTabs {
    constructor(public readonly underlying: Tabs, public readonly fsms: FuzzyTabFSM[]) {}
    public activeIndex(activeTab: FuzzyTabKey): number {
        return this.entries().findIndex(([key]) => activeTab === key)
    }
    public focusTabWithIncrement(activeTab: FuzzyTabKey, increment: number): FuzzyTabKey {
        const nextIndex = this.activeIndex(activeTab) + increment
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
            // eslint-disable-next-line @typescript-eslint/no-unsafe-member-access
            const value = (this.underlying as any)[key as keyof Tab] as Tab
            if (value.isEnabled) {
                result.push([key as FuzzyTabKey, value])
            }
        }
        return result
    }
    public all(): Tab[] {
        return Object.values(this.underlying).filter(tab => (tab as Tab).isEnabled)
    }
    public isOnlyFilesEnabled(): boolean {
        const [[tab], ...rest] = this.entries()
        return rest.length === 0 && tab === 'files'
    }
    public isDownloading(activeTab: FuzzyTabKey, scope: FuzzyScope): boolean {
        return this.fsms.find(tab => tab.isActive(activeTab, scope) && tab.fsm().key === 'downloading') !== undefined
    }
    public isAllDisabled(): boolean {
        return this.all().length === 0
    }
}

export function defaultFuzzyState(): FuzzyState {
    let query = ''
    let scope: FuzzyScope = 'repository'
    let activeTab: FuzzyTabKey = 'all'
    return {
        query,
        setQuery: newQuery => {
            query = newQuery
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
        tabs: new FuzzyTabs(defaultTabs, []),
        fsmGeneration: 0,
        scope,
        isScopeToggleDisabled: false,
        rankingCache: new SearchValueRankingCache(),
        setScope: newScope => {
            if (typeof newScope === 'function') {
                scope = newScope(scope)
            } else {
                scope = newScope
            }
        },
        toggleScope: () => {
            scope = scope === 'repository' ? 'everywhere' : 'repository'
        },
    }
}
export interface FuzzyTabsProps {
    settingsCascade: SettingsCascadeOrError<Settings>
    isRepositoryRelatedPage: boolean
    location: H.Location
    client?: ApolloClient<object>
    initialQuery?: string
    isVisible: boolean
    userHistory: UserHistory
    defaultActiveTab?: FuzzyTabKey
}

export function useFuzzyState(props: FuzzyTabsProps): FuzzyState {
    const {
        isVisible,
        location: { pathname, search, hash },
        isRepositoryRelatedPage,
        client: apolloClient,
        settingsCascade,
        userHistory,
        defaultActiveTab,
    } = props
    let {
        repoName = '',
        commitID = '',
        rawRevision = '',
    } = useMemo(() => {
        if (!isRepositoryRelatedPage) {
            return { repoName: '', commitID: '', rawRevision: '' }
        }
        return parseBrowserRepoURL(pathname + search + hash)
    }, [isRepositoryRelatedPage, pathname, search, hash])
    let revision = rawRevision || commitID
    if (!isRepositoryRelatedPage) {
        repoName = ''
        revision = ''
    }

    const repoRevision: FuzzyRepoRevision = useMemo(
        () => ({ repositoryName: repoName, revision }),
        [repoName, revision]
    )
    const repoRevisionRef = useRef<FuzzyRepoRevision>(repoRevision)
    repoRevisionRef.current = repoRevision

    const { fuzzyFinderAll, fuzzyFinderActions, fuzzyFinderRepositories, fuzzyFinderSymbols } =
        useFuzzyFinderFeatureFlags()

    const [activeTab, setActiveTab] = useState<FuzzyTabKey>(defaultActiveTab || 'all')

    // NOTE: the query is cached in session storage to mimic the file pickers in
    // IntelliJ (by default) and VS Code (when "Workbench > Quick Open >
    // Preserve Input" is enabled).
    const [query, setQuery] = useSessionStorage<string>(`fuzzy-modal.query.${repoName}`, props.initialQuery || '')

    // Re-initialize the cache whenever the query changes. We want to preserve
    // the ranking as long as the  queries to prevent jumpy ranking when the user
    // is cycling through results by repeatedly activating the fuzzy finder
    // eslint-disable-next-line react-hooks/exhaustive-deps
    const rankingCache = useMemo(() => new SearchValueRankingCache(), [query])

    const queryRef = useRef(query)
    queryRef.current = query

    // Scope determines whether to search for results within the repository of everywhere.
    const [scope, setScope] = useState<FuzzyScope>('repository')
    const toggleScope = useCallback(
        () => setScope(old => (old === 'repository' ? 'everywhere' : 'repository')),
        [setScope]
    )
    const isScopeToggleDisabled = activeTab === 'repos' || activeTab === 'actions' || !isRepositoryRelatedPage
    useEffect(() => {
        setScope(isScopeToggleDisabled ? 'everywhere' : 'repository')
    }, [isVisible, setScope, isScopeToggleDisabled])

    const [fsmGeneration, setFsmGeneration] = useState(0)
    const incrementFsmRenderGeneration: () => void = useCallback(
        () => setFsmGeneration(old => old + 1),
        [setFsmGeneration]
    )

    const createURL = useCallback(
        (filename: string): string =>
            toPrettyBlobURL({
                filePath: filename,
                revision,
                repoName,
            }),
        [revision, repoName]
    )

    const { theme, setThemeSetting } = useTheme()
    // Actions
    const actions = useMemo<FuzzyTabFSM>(() => {
        const fsm = createActionsFSM(getAllFuzzyActions({ theme, setThemeSetting }))
        return new FuzzyTabFSM('actions', 'always', () => fsm)
    }, [theme, setThemeSetting])

    // Repos
    const repos = useMemo<FuzzyTabFSM>(() => {
        const fsm = new FuzzyRepos(apolloClient, incrementFsmRenderGeneration, userHistory)
        return new FuzzyTabFSM(
            'repos',
            'everywhere',
            () => fsm.fuzzyFSM(),
            query => fsm.handleQuery(query)
        )
    }, [apolloClient, incrementFsmRenderGeneration, userHistory])

    // Symbols
    const localSymbols = useMemo<FuzzyTabFSM>(() => {
        const fsm = new FuzzySymbols(
            apolloClient,
            incrementFsmRenderGeneration,
            repoRevisionRef,
            false,
            settingsCascade,
            userHistory
        )
        return new FuzzyTabFSM(
            'symbols',
            'repository',
            () => fsm.fuzzyFSM(),
            query => fsm.handleQuery(query)
        )
    }, [apolloClient, incrementFsmRenderGeneration, settingsCascade, userHistory])
    const globalSymbols = useMemo<FuzzyTabFSM>(() => {
        const fsm = new FuzzySymbols(
            apolloClient,
            incrementFsmRenderGeneration,
            repoRevisionRef,
            true,
            settingsCascade,
            userHistory
        )
        return new FuzzyTabFSM(
            'symbols',
            'everywhere',
            () => fsm.fuzzyFSM(),
            query => fsm.handleQuery(query)
        )
    }, [apolloClient, incrementFsmRenderGeneration, settingsCascade, userHistory])

    // Files
    const localFiles = useMemo<FuzzyTabFSM>(() => {
        const fsm = new FuzzyRepoFiles(
            apolloClient,
            createURL,
            incrementFsmRenderGeneration,
            repoRevisionRef.current,
            userHistory
        )
        return new FuzzyTabFSM(
            'files',
            'repository',
            () => fsm.fuzzyFSM(),
            () => fsm.handleQuery()
        )
    }, [apolloClient, incrementFsmRenderGeneration, createURL, userHistory])
    const globalFiles = useMemo<FuzzyTabFSM>(() => {
        const fsm = new FuzzyFiles(apolloClient, incrementFsmRenderGeneration, repoRevisionRef, userHistory)
        return new FuzzyTabFSM(
            'files',
            'everywhere',
            () => fsm.fuzzyFSM(),
            query => fsm.handleQuery(query)
        )
    }, [apolloClient, incrementFsmRenderGeneration, userHistory])

    const tabs = useMemo(() => {
        const tabs: FuzzyTabFSM[] = []
        if (fuzzyFinderActions) {
            tabs.push(actions)
        }
        if (fuzzyFinderRepositories) {
            tabs.push(repos)
        }

        // Files are intentionally above symbols so that they rank above symbol results.
        tabs.push(localFiles)
        tabs.push(globalFiles)

        if (fuzzyFinderSymbols) {
            tabs.push(globalSymbols)
            tabs.push(localSymbols)
        }
        return new FuzzyTabs(
            {
                all: fuzzyFinderAll ? defaultTabs.all : hiddenKind,
                actions: fuzzyFinderActions ? defaultTabs.actions : hiddenKind,
                repos: fuzzyFinderRepositories ? defaultTabs.repos : hiddenKind,
                symbols: fuzzyFinderSymbols ? defaultTabs.symbols : hiddenKind,
                files: defaultTabs.files,
                lines: hiddenKind,
            },
            tabs
        )
    }, [
        fuzzyFinderAll,
        fuzzyFinderActions,
        fuzzyFinderRepositories,
        fuzzyFinderSymbols,
        actions,
        repos,
        globalFiles,
        localFiles,
        globalSymbols,
        localSymbols,
    ])

    return {
        query,
        setQuery,
        activeTab,
        setActiveTab,
        repoRevision,
        fsmGeneration,
        scope,
        toggleScope,
        setScope,
        isScopeToggleDisabled,
        tabs,
        rankingCache,
    }
}
