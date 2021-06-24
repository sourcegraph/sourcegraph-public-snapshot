import { storiesOf } from '@storybook/react'
import { noop } from 'lodash'
import React from 'react'
import { EMPTY, NEVER, of } from 'rxjs'

import { FlatExtensionHostAPI } from '@sourcegraph/shared/src/api/contract'
import { PanelViewData } from '@sourcegraph/shared/src/api/extension/extensionHostApi'
import { pretendProxySubscribable, pretendRemote } from '@sourcegraph/shared/src/api/util'
import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { extensionsController } from '@sourcegraph/shared/src/util/searchTestHelpers'
import webStyles from '@sourcegraph/web/src/SourcegraphWebApp.scss'

import { BrandedStory } from '../BrandedStory'

import { Panel } from './Panel'

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

const { add } = storiesOf('branded/Panel', module)
    .addDecorator(story => (
        <BrandedStory styles={webStyles} initialEntries={[{ pathname: '/', hash: `#tab=${panels[0].id}` }]}>
            {() => <div className="p-4">{story()}</div>}
        </BrandedStory>
    ))
    .addParameters({
        chromatic: {
            viewports: [320, 576, 978, 1440],
        },
    })

const panelActions = [
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

const panelMenus = {
    'panel/toolbar': [
        {
            action: 'a',
        },
        {
            action: 'b',
        },
    ],
}

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

add('Simple', () => <Panel {...panelProps} />)

add('With actions', () => (
    <Panel
        {...panelProps}
        extensionsController={{
            ...extensionsController,
            extHostAPI: Promise.resolve(
                pretendRemote<FlatExtensionHostAPI>({
                    getContributions: () => pretendProxySubscribable(of({ actions: panelActions, menus: panelMenus })),
                    registerContributions: () => pretendProxySubscribable(EMPTY).subscribe(noop as any),
                    haveInitialExtensionsLoaded: () => pretendProxySubscribable(of(true)),
                    getPanelViews: () => pretendProxySubscribable(of(panels)),
                    getActiveCodeEditorPosition: () => pretendProxySubscribable(of(null)),
                })
            ),
        }}
    />
))
