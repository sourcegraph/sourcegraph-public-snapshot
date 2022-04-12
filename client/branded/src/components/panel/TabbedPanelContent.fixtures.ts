import { noop } from 'lodash'
import { EMPTY, NEVER, of } from 'rxjs'

import { FlatExtensionHostAPI } from '@sourcegraph/shared/src/api/contract'
import { ExtensionCodeEditor } from '@sourcegraph/shared/src/api/extension/api/codeEditor'
import { ExtensionDocument } from '@sourcegraph/shared/src/api/extension/api/textDocument'
import { PanelViewData } from '@sourcegraph/shared/src/api/extension/extensionHostApi'
import { pretendProxySubscribable, pretendRemote } from '@sourcegraph/shared/src/api/util'
import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { extensionsController } from '@sourcegraph/shared/src/testing/searchTestHelpers'

export const panels: PanelViewData[] = [
    {
        id: 'panel_1',
        title: 'Panel 1',
        content: 'Panel 1',
        priority: 3,
        component: null,
        selector: null,
    },
    {
        id: 'panel_2',
        title: 'Panel 2',
        content: 'Panel 2',
        priority: 2,
        component: null,
        selector: null,
    },
    {
        id: 'panel_3',
        title: 'Panel 3',
        content: 'Panel 3',
        priority: 1,
        component: null,
        selector: null,
    },
]

export const panelActions = [
    {
        id: 'a',
        actionItem: {
            label: 'Action A',
            description: 'This is Action A',
        },
        command: 'open',
        commandArguments: ['https://example.com'],
    },
    {
        id: 'b',
        actionItem: {
            label: 'Action B',
            description: 'This is Action B',
        },
        command: 'updateConfiguration',
        commandArguments: [],
    },
]

export const panelMenus = {
    'panel/toolbar': [
        {
            action: 'a',
        },
        {
            action: 'b',
        },
    ],
}

export const panelProps = {
    repoName: 'git://github.com/foo/bar',
    fetchHighlightedFileLineRanges: () => of([]),
    isLightTheme: true,
    platformContext: {} as any,
    settingsCascade: { subjects: null, final: null },
    telemetryService: NOOP_TELEMETRY_SERVICE,
    extensionsController: {
        ...extensionsController,
        extHostAPI: Promise.resolve(
            pretendRemote<FlatExtensionHostAPI>({
                getContributions: () => pretendProxySubscribable(NEVER),
                registerContributions: () => pretendProxySubscribable(EMPTY).subscribe(noop as any),
                haveInitialExtensionsLoaded: () => pretendProxySubscribable(of(true)),
                getPanelViews: () => pretendProxySubscribable(of(panels)),
                getActiveViewComponentChanges: () => pretendProxySubscribable(of(CODE_EDITOR_FIXTURE)),
                getActiveCodeEditorPosition: () => pretendProxySubscribable(NEVER),
            })
        ),
    },
}

export const CODE_EDITOR_FIXTURE = new ExtensionCodeEditor(
    {
        type: 'CodeEditor',
        viewerId: 'viewer#0',
        resource: 'git://foo?1#/bar.go',
        selections: [],
        isActive: true,
    },
    new ExtensionDocument({
        uri: 'git://foo?1#/bar.go',
        languageId: 'go',
        text: 'type My[Kingdom For] Generics',
    })
)
