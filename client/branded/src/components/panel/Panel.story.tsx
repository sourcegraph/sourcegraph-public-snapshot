import { noop } from 'lodash'
import React from 'react'
import { EMPTY, of } from 'rxjs'

import { FlatExtensionHostAPI } from '@sourcegraph/shared/src/api/contract'
import { pretendProxySubscribable, pretendRemote } from '@sourcegraph/shared/src/api/util'
import { extensionsController } from '@sourcegraph/shared/src/util/searchTestHelpers'
import webStyles from '@sourcegraph/web/src/SourcegraphWebApp.scss'

import { BrandedStory } from '../BrandedStory'

import { Panel } from './Panel'
import { panels, panelProps, panelActions, panelMenus } from './Panel.fixtures'

export default {
    title: 'branded/Panel',

    decorators: [
        story => (
            <BrandedStory styles={webStyles} initialEntries={[{ pathname: '/', hash: `#tab=${panels[0].id}` }]}>
                {() => <div className="p-4">{story()}</div>}
            </BrandedStory>
        ),
    ],

    parameters: {
        chromatic: {
            viewports: [320, 576, 978, 1440],
        },
    },
}

export const Simple = () => <Panel {...panelProps} />

export const WithActions = () => (
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
)

WithActions.story = {
    name: 'With actions',
}
