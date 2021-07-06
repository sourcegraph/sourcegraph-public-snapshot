import { storiesOf } from '@storybook/react'
import React from 'react'

import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { WebStory } from '../../components/WebStory'

import { RepogroupPanel } from './RepogroupPanel'
import { _fetchRecentSearches } from './utils'

const { add } = storiesOf('web/search/panels/RepogroupPanel', module)
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

add('Populated', () => <WebStory>{() => <RepogroupPanel {...props} />}</WebStory>)
