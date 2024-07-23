import type { Decorator, Meta, StoryFn } from '@storybook/react'

import { BrandedStory } from '@sourcegraph/wildcard/src/stories'

import { TabbedPanelContent } from './TabbedPanelContent'
import { panels, panelProps } from './TabbedPanelContent.fixtures'

const decorator: Decorator = story => (
    <BrandedStory initialEntries={[{ pathname: '/', hash: `#tab=${panels[0].id}` }]}>
        {() => <div className="p-4">{story()}</div>}
    </BrandedStory>
)
const config: Meta = {
    title: 'branded/TabbedPanelContent',
    decorators: [decorator],
    parameters: {},
}

export default config

export const Simple: StoryFn = () => <TabbedPanelContent {...panelProps} />

Simple.storyName = 'Simple'
