import * as comlink from 'comlink'
import { BehaviorSubject, Observable, of, ReplaySubject, Subject } from 'rxjs'
import * as sourcegraph from 'sourcegraph'

import { Contributions } from '@sourcegraph/client-api'
import { isErrorLike } from '@sourcegraph/common'
import { Context } from '@sourcegraph/template-parser'

import { ConfiguredExtension } from '../../extensions/extension'
import { Settings, SettingsCascade } from '../../settings/settings'
import { MainThreadAPI } from '../contract'
import { ExtensionViewer, ViewerUpdate } from '../viewerTypes'

import { ExecutableExtension, observeActiveExtensions } from './activation'
import { ExtensionCodeEditor } from './api/codeEditor'
import { ExtensionDocument } from './api/textDocument'
import { ExtensionWorkspaceRoot } from './api/workspaceRoot'
import { InitData } from './extensionHost'
import {
    RegisteredProvider,
    RegisteredViewProvider,
    PanelViewData,
    PlainNotification,
    ProgressNotification,
} from './extensionHostApi'
import { ReferenceCounter } from './utils/ReferenceCounter'

export function createExtensionHostState(
    initData: Pick<InitData, 'initialSettings' | 'clientApplication'>,
    mainAPI: comlink.Remote<MainThreadAPI> | null,
    mainThreadAPIInitializations: Observable<boolean> | null
): ExtensionHostState {
    // We make the mainAPI nullable in which case no extension will ever be activated. This is
    // used only for the noop controller.
    let activeLanguages = new BehaviorSubject<ReadonlySet<string>>(new Set())
    let activeExtensions: Observable<(ConfiguredExtension | ExecutableExtension)[]> = of([])
    if (mainAPI !== null && mainThreadAPIInitializations !== null) {
        ;({ activeLanguages, activeExtensions } = observeActiveExtensions(mainAPI, mainThreadAPIInitializations))
    }

    return {
        haveInitialExtensionsLoaded: new BehaviorSubject<boolean>(false),

        roots: new BehaviorSubject<readonly ExtensionWorkspaceRoot[]>([]),
        rootChanges: new Subject<void>(),
        searchContextChanges: new Subject<string | undefined>(),
        searchContext: undefined,

        // Most extensions never call `configuration.get()` synchronously in `activate()` to get
        // the initial settings data, and instead only subscribe to configuration changes.
        // In order for these extensions to be able to access settings, make sure `configuration` emits on subscription.
        settings: new BehaviorSubject<Readonly<SettingsCascade>>(initData.initialSettings),

        queryTransformers: new BehaviorSubject<readonly sourcegraph.QueryTransformer[]>([]),

        hoverProviders: new BehaviorSubject<readonly RegisteredProvider<sourcegraph.HoverProvider>[]>([]),
        documentHighlightProviders: new BehaviorSubject<
            readonly RegisteredProvider<sourcegraph.DocumentHighlightProvider>[]
        >([]),
        definitionProviders: new BehaviorSubject<readonly RegisteredProvider<sourcegraph.DefinitionProvider>[]>([]),
        referenceProviders: new BehaviorSubject<readonly RegisteredProvider<sourcegraph.ReferenceProvider>[]>([]),
        locationProviders: new BehaviorSubject<
            readonly RegisteredProvider<{ id: string; provider: sourcegraph.LocationProvider }>[]
        >([]),

        fileDecorationProviders: new BehaviorSubject<readonly sourcegraph.FileDecorationProvider[]>([]),

        context: new BehaviorSubject<Context>({
            'clientApplication.isSourcegraph': initData.clientApplication === 'sourcegraph',

            'experimentalFeatures.enableExtensionsDecorationsColumnView':
                !isErrorLike(initData.initialSettings.final) &&
                (initData.initialSettings.final as Settings).experimentalFeatures
                    ?.enableExtensionsDecorationsColumnView === true,

            // Arbitrary, undocumented versioning for extensions that need different behavior for different
            // Sourcegraph versions.
            //
            // TODO: Make this more advanced if many extensions need this (although we should try to avoid
            // extensions needing this).
            'clientApplication.extensionAPIVersion.major': 3,
        }),
        contributions: new BehaviorSubject<readonly Contributions[]>([]),

        lastViewerId: 0,
        textDocuments: new Map<string, ExtensionDocument>(),
        openedTextDocuments: new Subject<ExtensionDocument>(),
        viewComponents: new Map<string, ExtensionCodeEditor>(),

        activeLanguages,
        languageReferences: new ReferenceCounter<string>(),
        modelReferences: new ReferenceCounter<string>(),

        activeViewComponentChanges: new BehaviorSubject<ExtensionViewer | undefined>(undefined),
        viewerUpdates: new Subject<ViewerUpdate>(),

        // Use ReplaySubject so we don't lose notifications in case the client application subscribes
        // to notification streams after extensions have already sent notifications.
        // There should be no issue re: stale notifications, since client applications should only
        // create one "notification manager" instance.
        plainNotifications: new ReplaySubject<PlainNotification>(3),
        progressNotifications: new ReplaySubject<ProgressNotification & comlink.ProxyMarked>(3),

        panelViews: new BehaviorSubject<readonly Observable<PanelViewData>[]>([]),
        insightsPageViewProviders: new BehaviorSubject<readonly RegisteredViewProvider<'insightsPage'>[]>([]),
        homepageViewProviders: new BehaviorSubject<readonly RegisteredViewProvider<'homepage'>[]>([]),
        globalPageViewProviders: new BehaviorSubject<readonly RegisteredViewProvider<'global/page'>[]>([]),
        directoryViewProviders: new BehaviorSubject<readonly RegisteredViewProvider<'directory'>[]>([]),

        linkPreviewProviders: new BehaviorSubject<
            readonly { urlMatchPattern: string; provider: sourcegraph.LinkPreviewProvider }[]
        >([]),

        activeExtensions,
        activeLoggers: new Set<string>(),
    }
}

export interface ExtensionHostState {
    haveInitialExtensionsLoaded: BehaviorSubject<boolean>
    settings: BehaviorSubject<Readonly<SettingsCascade>>

    // Workspace
    roots: BehaviorSubject<readonly ExtensionWorkspaceRoot[]>
    rootChanges: Subject<void>
    searchContextChanges: Subject<string | undefined>
    searchContext: string | undefined

    // Search
    queryTransformers: BehaviorSubject<readonly sourcegraph.QueryTransformer[]>

    // Language features
    hoverProviders: BehaviorSubject<readonly RegisteredProvider<sourcegraph.HoverProvider>[]>
    documentHighlightProviders: BehaviorSubject<readonly RegisteredProvider<sourcegraph.DocumentHighlightProvider>[]>
    definitionProviders: BehaviorSubject<readonly RegisteredProvider<sourcegraph.DefinitionProvider>[]>
    referenceProviders: BehaviorSubject<readonly RegisteredProvider<sourcegraph.ReferenceProvider>[]>
    locationProviders: BehaviorSubject<
        readonly RegisteredProvider<{ id: string; provider: sourcegraph.LocationProvider }>[]
    >

    // Decorations
    fileDecorationProviders: BehaviorSubject<readonly sourcegraph.FileDecorationProvider[]>

    // Context + Contributions
    context: BehaviorSubject<Context>
    /** All contributions, including those that are not enabled in the current context. */
    contributions: BehaviorSubject<readonly Contributions[]>

    // Viewer + Text documents
    lastViewerId: number
    openedTextDocuments: Subject<ExtensionDocument>
    activeLanguages: BehaviorSubject<ReadonlySet<string>>
    modelReferences: ReferenceCounter<string>
    languageReferences: ReferenceCounter<string>
    /** Mutable map of URIs to text documents */
    textDocuments: Map<string, ExtensionDocument>

    /** Mutable map of viewer ID to viewer. */
    viewComponents: Map<string, ExtensionViewer>
    activeViewComponentChanges: BehaviorSubject<ExtensionViewer | undefined>
    viewerUpdates: Subject<ViewerUpdate>

    plainNotifications: ReplaySubject<PlainNotification>
    progressNotifications: ReplaySubject<ProgressNotification & comlink.ProxyMarked>

    // Views
    panelViews: BehaviorSubject<readonly Observable<PanelViewData>[]>
    insightsPageViewProviders: BehaviorSubject<readonly RegisteredViewProvider<'insightsPage'>[]>
    homepageViewProviders: BehaviorSubject<readonly RegisteredViewProvider<'homepage'>[]>
    globalPageViewProviders: BehaviorSubject<readonly RegisteredViewProvider<'global/page'>[]>
    directoryViewProviders: BehaviorSubject<readonly RegisteredViewProvider<'directory'>[]>

    // Content
    linkPreviewProviders: BehaviorSubject<
        readonly { urlMatchPattern: string; provider: sourcegraph.LinkPreviewProvider }[]
    >

    // Extensions
    activeExtensions: Observable<(ConfiguredExtension | ExecutableExtension)[]>

    /** Set of names of active loggers determined by user settings */
    activeLoggers: Set<string>
}
