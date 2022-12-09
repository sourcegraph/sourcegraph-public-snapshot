import { DecoratorFn, Meta, Story } from '@storybook/react'

import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { WebStory } from '../../components/WebStory'

import { CommunitySearchContextsPanel } from './CommunitySearchContextPanel'

const decorator: DecoratorFn = story => <div style={{ width: '800px' }}>{story()}</div>

const config: Meta = {
    title: 'web/search/panels/CommunitySearchContextPanel',
    parameters: {
        design: {
            type: 'figma',
            url: 'https://www.figma.com/file/zCGglxWBFm8Fv5DwdcHdAQ/Repository-group-home-page-panel-14393?node-id=1%3A159',
        },
        chromatic: { viewports: [800] },
    },
    decorators: [decorator],
}

export default config

const props = {
    authenticatedUser: null,
    telemetryService: NOOP_TELEMETRY_SERVICE,
}

export const CommunitySearchContextPanelStory: Story = () => (
    <WebStory>
        {() => (
            <div style={{ maxWidth: '32rem' }}>
                <CommunitySearchContextsPanel {...props} />
            </div>
        )}
    </WebStory>
)

CommunitySearchContextPanelStory.storyName = 'CommunitySearchContextsPanel'
