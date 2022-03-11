import { Meta, DecoratorFn } from '@storybook/react'
import React from 'react'

import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { WebStory } from '../components/WebStory'

import { GettingStartedTour } from './GettingStartedTour'

const decorator: DecoratorFn = story => <WebStory>{() => <div className="container mt-3">{story()}</div>}</WebStory>

const config: Meta = {
    title: 'web/GettingStartedTour',
    decorators: [decorator],
    parameters: {
        component: GettingStartedTour,
        chromatic: {
            enableDarkMode: true,
            disableSnapshot: false,
        },
    },
}

export default config

export const Default: React.FunctionComponent = () => <GettingStartedTour telemetryService={NOOP_TELEMETRY_SERVICE} />
