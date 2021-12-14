import { Meta } from '@storybook/react'
import React from 'react'

import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { WebStory } from '../../../../../../components/WebStory'

import { IntroCreationPage } from './IntroCreationPage'

export default {
    title: 'web/insights/creation-ui/IntroPage',
    decorators: [story => <WebStory>{() => story()}</WebStory>],
    parameters: {
        chromatic: {
            viewports: [576, 978, 1440],
        },
    },
} as Meta

export const InsightIntroPageExample = () => <IntroCreationPage telemetryService={NOOP_TELEMETRY_SERVICE} />
