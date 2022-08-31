import { NEVER, of } from 'rxjs'

import { FlatExtensionHostAPI } from '../api/contract'
import { ProxySubscribable, proxySubscribable } from '../api/extension/api/common'
import { pretendRemote } from '../api/util'

import { Controller } from './controller'

export function createNoopController(): Controller {
    return {
        executeCommand: () => Promise.resolve(),
        commandErrors: NEVER,
        registerCommand: () => ({
            unsubscribe: () => {},
        }),
        extHostAPI: Promise.resolve(pretendRemote(noopFlatExtensionHostAPI)),
        unsubscribe: () => {},
    }
}

const NOOP = (): void => {}
// eslint-disable-next-line @typescript-eslint/no-explicit-any
const NOOP_EMPTY_ARRAY_PROXY = (): ProxySubscribable<any> => proxySubscribable(of([]))
// eslint-disable-next-line @typescript-eslint/no-explicit-any
const NOOP_NEVER_PROXY = (): ProxySubscribable<any> => proxySubscribable(NEVER)

const noopFlatExtensionHostAPI: FlatExtensionHostAPI = {
    syncSettingsData: NOOP,

    addWorkspaceRoot: NOOP,
    getWorkspaceRoots: NOOP_EMPTY_ARRAY_PROXY,
    removeWorkspaceRoot: NOOP,

    setSearchContext: NOOP,
    transformSearchQuery: (query: string) => proxySubscribable(of(query)),

    getHover: () => proxySubscribable(of({ isLoading: true, result: null })),
    getDocumentHighlights: NOOP_EMPTY_ARRAY_PROXY,
    getDefinition: () => proxySubscribable(of({ isLoading: true, result: [] })),
    getReferences: () => proxySubscribable(of({ isLoading: true, result: [] })),
    getLocations: () => proxySubscribable(of({ isLoading: true, result: [] })),

    hasReferenceProvidersForDocument: () => proxySubscribable(of(false)),

    getFileDecorations: () => proxySubscribable(of({})),

    updateContext: NOOP,

    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    registerContributions: (): any => ({
        unsubscribe: NOOP,
    }),
    getContributions: () => proxySubscribable(of({})),

    addTextDocumentIfNotExists: NOOP,

    getActiveViewComponentChanges: () => proxySubscribable(of(undefined)),

    getActiveCodeEditorPosition: () => proxySubscribable(of(null)),

    getTextDecorations: NOOP_EMPTY_ARRAY_PROXY,

    addViewerIfNotExists: () => ({ viewerId: '' }),
    viewerUpdates: NOOP_NEVER_PROXY,

    setEditorSelections: NOOP,
    removeViewer: NOOP,

    getPlainNotifications: NOOP_NEVER_PROXY,
    getProgressNotifications: NOOP_NEVER_PROXY,

    getPanelViews: NOOP_EMPTY_ARRAY_PROXY,

    getInsightViewById: NOOP_NEVER_PROXY,
    getInsightsViews: NOOP_EMPTY_ARRAY_PROXY,

    getHomepageViews: NOOP_EMPTY_ARRAY_PROXY,

    getDirectoryViews: NOOP_EMPTY_ARRAY_PROXY,

    getGlobalPageViews: NOOP_EMPTY_ARRAY_PROXY,
    getStatusBarItems: NOOP_EMPTY_ARRAY_PROXY,

    getLinkPreviews: () => proxySubscribable(of(null)),

    haveInitialExtensionsLoaded: () => proxySubscribable(of(true)),

    getActiveExtensions: NOOP_EMPTY_ARRAY_PROXY,
}
