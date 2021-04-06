import { storiesOf } from '@storybook/react'
import React from 'react'
import { MemoryRouter } from 'react-router'
import { EMPTY, NEVER, of } from 'rxjs'
import { noop } from 'lodash'
import webStyles from '../../../../web/src/main.scss'
import { Panel } from './Panel'
import { extensionsController } from '../../../../shared/src/util/searchTestHelpers'
import { NOOP_TELEMETRY_SERVICE } from '../../../../shared/src/telemetry/telemetryService'
import { pretendProxySubscribable, pretendRemote } from '../../../../shared/src/api/util'
import { FlatExtensionHostAPI } from '../../../../shared/src/api/contract'
import { PanelViewData } from '../../../../shared/src/api/extension/extensionHostApi'

const { add } = storiesOf('branded/Panel', module).addDecorator(story => (
    <>
        <div className="p-4">{story()}</div>
        <style>{webStyles}</style>
    </>
))

const panels: PanelViewData[] = [
    {
        id: 'panel_1',
        title: 'Panel 1',
        content: 'Panel 1',
        priority: 3,
        component: null,
    },
    {
        id: 'panel_2',
        title: 'Panel 2',
        content: 'Panel 2',
        priority: 2,
        component: null,
    },
    {
        id: 'panel_3',
        title: 'Panel 3',
        content: 'Panel 3',
        priority: 1,
        component: null,
    },
]

const panelProps = {
    repoName: 'git://github.com/foo/bar',
    fetchHighlightedFileLineRanges: () => of([]),
    isLightTheme: true,
    versionContext: undefined,
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
                getActiveCodeEditorPosition: () => pretendProxySubscribable(NEVER),
            })
        ),
    },
}

const PanelWithActions = (
    <Panel
        {...panelProps}
        extensionsController={{
            ...extensionsController,
            extHostAPI: Promise.resolve(
                pretendRemote<FlatExtensionHostAPI>({
                    getContributions: () =>
                        pretendProxySubscribable(
                            of({
                                actions: [
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
                                ],
                                menus: {
                                    'panel/toolbar': [
                                        {
                                            action: 'a',
                                        },
                                        {
                                            action: 'b',
                                        },
                                    ],
                                },
                            })
                        ),
                    registerContributions: () => pretendProxySubscribable(EMPTY).subscribe(noop as any),
                    haveInitialExtensionsLoaded: () => pretendProxySubscribable(of(true)),
                    getPanelViews: () => pretendProxySubscribable(of(panels)),
                    getActiveCodeEditorPosition: () => pretendProxySubscribable(of(null)),
                })
            ),
        }}
    />
)

add('Simple', () => (
    <MemoryRouter initialEntries={[{ pathname: '/', hash: `#tab=${panels[0].id}` }]}>
        <div>
            <Panel {...panelProps} />
        </div>
    </MemoryRouter>
))

add('Simple mobile', () => (
    <MemoryRouter initialEntries={[{ pathname: '/', hash: `#tab=${panels[0].id}` }]}>
        <div style={{ width: 320 }}>
            <Panel {...panelProps} />
        </div>
    </MemoryRouter>
))

add('With actions', () => (
    <MemoryRouter initialEntries={[{ pathname: '/', hash: `#tab=${panels[0].id}` }]}>
        <div>{PanelWithActions}</div>
    </MemoryRouter>
))

add('With actions mobile', () => (
    <MemoryRouter initialEntries={[{ pathname: '/', hash: `#tab=${panels[0].id}` }]}>
        <div style={{ width: 320 }}>{PanelWithActions}</div>
    </MemoryRouter>
))
