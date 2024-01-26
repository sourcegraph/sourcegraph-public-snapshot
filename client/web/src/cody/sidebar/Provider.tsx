import React, { useContext, useState, useCallback, useMemo } from 'react'

import ignore from 'ignore'

import { useQuery, gql } from '@sourcegraph/http-client'
import type { AuthenticatedUser } from '@sourcegraph/shared/src/auth'
import { useTemporarySetting } from '@sourcegraph/shared/src/settings/temporary'

import { parseBrowserRepoURL } from '../../util/url'
import { useCodyChat, type CodyChatStore, codyChatStoreMock } from '../useCodyChat'

import { useSidebarSize } from './useSidebarSize'

interface CodySidebarStore extends CodyChatStore {
    readonly isSidebarOpen: boolean
    readonly inputNeedsFocus: boolean
    setIsSidebarOpen: (isOpen: boolean) => void
    setFocusProvided: () => void
    setSidebarSize: (size: number) => void
    ignores: (path: string) => boolean
}

const CodySidebarContext = React.createContext<CodySidebarStore | null>({
    ...codyChatStoreMock,
    isSidebarOpen: false,
    inputNeedsFocus: false,
    setSidebarSize: () => {},
    setIsSidebarOpen: () => {},
    setFocusProvided: () => {},
    ignores: () => false,
})

interface ICodySidebarStoreProviderProps {
    children?: React.ReactNode
    authenticatedUser: AuthenticatedUser | null
}

const MY_QUERY = gql`
    query CodyIgnoreContent($repoName: String!, $repoRev: String!, $filePath: String!) {
        repository(name: $repoName) {
            commit(rev: $repoRev) {
                blob(path: $filePath) {
                    content
                }
            }
        }
    }
`

const CODY_IGNORE_PATH = '.cody/ignore'
const useCodyIgnore = (): { ignores: (path: string) => boolean } => {
    const { repoName, revision } = parseBrowserRepoURL(location.pathname + location.search + location.hash)
    const { data } = useQuery<any, any>(MY_QUERY, {
        variables: { repoName, repoRev: revision, filePath: CODY_IGNORE_PATH },
    })
    const ignoreManager = useMemo(() => (data ? ignore().add(data.repository.commit.blob.content || '') : null), [data])

    // TODO: remove
    window.ignoreManager = ignoreManager

    const ignores = useCallback(
        (path: string): boolean => {
            if (ignoreManager) {
                return ignoreManager.ignores(path)
            }
            return false
        },
        [ignoreManager]
    )
    return { ignores }
}

export const CodySidebarStoreProvider: React.FC<ICodySidebarStoreProviderProps> = ({ authenticatedUser, children }) => {
    const { ignores } = useCodyIgnore()
    const [isSidebarOpen, setIsSidebarOpenState] = useTemporarySetting('cody.showSidebar', false)
    const [inputNeedsFocus, setInputNeedsFocus] = useState(false)
    const { setSidebarSize } = useSidebarSize()

    const setFocusProvided = useCallback(() => {
        setInputNeedsFocus(false)
    }, [setInputNeedsFocus])

    const setIsSidebarOpen = useCallback(
        (open: boolean) => {
            setIsSidebarOpenState(open)
            setInputNeedsFocus(true)
        },
        [setIsSidebarOpenState, setInputNeedsFocus]
    )

    const onEvent = useCallback(() => setIsSidebarOpen(true), [setIsSidebarOpen])

    const codyChatStore = useCodyChat({ userID: authenticatedUser?.id, onEvent })

    const state = useMemo<CodySidebarStore>(
        () => ({
            ...codyChatStore,
            isSidebarOpen: isSidebarOpen ?? false,
            inputNeedsFocus,
            setIsSidebarOpen,
            setFocusProvided,
            setSidebarSize,
            ignores,
        }),
        [codyChatStore, isSidebarOpen, setIsSidebarOpen, setFocusProvided, setSidebarSize, inputNeedsFocus, ignores]
    )

    // dirty fix because CodyRecipesWidget is rendered inside a different React DOM tree.
    const global = window as any
    global.codySidebarStore = state

    return <CodySidebarContext.Provider value={state}>{children}</CodySidebarContext.Provider>
}

export const useCodySidebar = (): CodySidebarStore => useContext(CodySidebarContext) as CodySidebarStore

export const CODY_SIDEBAR_SIZES = { default: 350, max: 1200, min: 250 }

export { useSidebarSize }
