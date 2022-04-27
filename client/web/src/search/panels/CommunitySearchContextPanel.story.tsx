import { storiesOf } from '@storybook/react'

import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { WebStory } from '../../components/WebStory'

import { CommunitySearchContextsPanel } from './CommunitySearchContextPanel'

const { add } = storiesOf('web/search/panels/CommunitySearchContextPanel', module)
    .addParameters({
        design: {
            type: 'figma',
            url:
                'https://www.figma.com/file/zCGglxWBFm8Fv5DwdcHdAQ/Repository-group-home-page-panel-14393?node-id=1%3A159',
        },
        chromatic: { viewports: [800] },
    })
    .addDecorator(story => <div style={{ width: '800px' }}>{story()}</div>)

const props = {
    authenticatedUser: null,
    telemetryService: NOOP_TELEMETRY_SERVICE,
}

add('CommunitySearchContextPanel', () => (
    <WebStory>
        {() => (
            <div style={{ maxWidth: '32rem' }}>
                <CommunitySearchContextsPanel {...props} />
            </div>
        )}
    </WebStory>
))
