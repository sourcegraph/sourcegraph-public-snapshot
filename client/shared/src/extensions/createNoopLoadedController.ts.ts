import { Remote } from 'comlink'
import { NEVER, of } from 'rxjs'

import { FlatExtensionHostAPI } from '../api/contract'
import { ProxySubscribable, proxySubscribable } from '../api/extension/api/common'

import { Controller } from './controller'

export function createNoopController(): Controller {
    return {
        executeCommand: () => Promise.resolve(),
        commandErrors: NEVER,
        registerCommand: () => ({
            unsubscribe: () => {},
        }),
        extHostAPI: Promise.resolve(makeRemote(noopFlatExtensionHostAPI)),
        unsubscribe: () => {},
    }
}

const NOOP = (): void => {}
const NOOP_EMPTY_ARRAY_PROXY = (): ProxySubscribable<any> => proxySubscribable(of([]))
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

// Since our public Controller interface exposes the APIs on the comlink internal Remote<> type, we need
// to mimic this behavior. A remote oject is an object where all methods are wrapped in a Promise. We
// use a proxy to achieve this.
function makeRemote<T>(object: T): Remote<T> {
    const proxy: any = new Proxy(
        {},
        {
            get(_target, prop) {
                const raw = object[prop as keyof T]
                if (typeof raw === 'function') {
                    return (...args: any[]) => Promise.resolve(raw(...args))
                }
                return Promise.resolve(raw)
            },
            set() {
                throw new Error('set is not supported on the extensions controller')
            },
            apply() {
                throw new Error('apply is not supported on the extensions controller')
            },
            construct() {
                throw new Error('construct is not supported on the extensions controller')
            },
        }
    )
    return proxy
}
