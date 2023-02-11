import { DecoratorFn, Meta, Story } from '@storybook/react'

import { BrandedStory } from '@sourcegraph/wildcard/src/stories'

import { TabbedPanelContent } from './TabbedPanelContent'
import { panels, panelProps } from './TabbedPanelContent.fixtures'

const decorator: DecoratorFn = story => (
    <BrandedStory initialEntries={[{ pathname: '/', hash: `#tab=${panels[0].id}` }]}>
        {() => <div className="p-4">{story()}</div>}
    </BrandedStory>
)
const config: Meta = {
    title: 'branded/TabbedPanelContent',
    decorators: [decorator],
    parameters: {
        chromatic: {
            viewports: [320, 576, 978, 1440],
        },
    },
}

export default config

export const Simple: Story = () => <TabbedPanelContent {...panelProps} />

Simple.storyName = 'Simple'
