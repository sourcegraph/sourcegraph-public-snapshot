import { DecoratorFn, Meta, Story } from '@storybook/react'
import { noop } from 'lodash'
import React from 'react'
import { EMPTY, of } from 'rxjs'

import { FlatExtensionHostAPI } from '@sourcegraph/shared/src/api/contract'
import { pretendProxySubscribable, pretendRemote } from '@sourcegraph/shared/src/api/util'
import { extensionsController } from '@sourcegraph/shared/src/testing/searchTestHelpers'
import webStyles from '@sourcegraph/web/src/SourcegraphWebApp.scss'

import { BrandedStory } from '../BrandedStory'

import { TabbedPanel } from './TabbedPanel'
import { panels, panelProps, panelActions, panelMenus, CODE_EDITOR_FIXTURE } from './TabbedPanel.fixtures'

const decorator: DecoratorFn = story => (
    <BrandedStory styles={webStyles} initialEntries={[{ pathname: '/', hash: `#tab=${panels[0].id}` }]}>
        {() => <div className="p-4">{story()}</div>}
    </BrandedStory>
)
const config: Meta = {
    title: 'branded/TabbedPanel',
    decorators: [decorator],
    parameters: {
        chromatic: {
            viewports: [320, 576, 978, 1440],
        },
    },
}

export default config

export const Simple: Story = () => <TabbedPanel {...panelProps} />

export const WithActions: Story = () => (
    <TabbedPanel
        {...panelProps}
        extensionsController={{
            ...extensionsController,
            extHostAPI: Promise.resolve(
                pretendRemote<FlatExtensionHostAPI>({
                    getContributions: () => pretendProxySubscribable(of({ actions: panelActions, menus: panelMenus })),
                    registerContributions: () => pretendProxySubscribable(EMPTY).subscribe(noop as any),
                    haveInitialExtensionsLoaded: () => pretendProxySubscribable(of(true)),
                    getPanelViews: () => pretendProxySubscribable(of(panels)),
                    getActiveViewComponentChanges: () => pretendProxySubscribable(of(CODE_EDITOR_FIXTURE)),
                    getActiveCodeEditorPosition: () => pretendProxySubscribable(of(null)),
                })
            ),
        }}
    />
)

WithActions.storyName = 'With actions'
