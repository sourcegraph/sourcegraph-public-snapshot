import { Meta, DecoratorFn } from '@storybook/react'
import React from 'react'

import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { WebStory } from '../components/WebStory'

import { OnboardingTour } from './OnboardingTour'

const decorator: DecoratorFn = story => <WebStory>{() => <div className="container mt-3">{story()}</div>}</WebStory>

const config: Meta = {
    title: 'web/OnboardingTour',
    decorators: [decorator],
    parameters: {
        component: OnboardingTour,
    },
}

export default config

export const Default: React.FunctionComponent = () => <OnboardingTour telemetryService={NOOP_TELEMETRY_SERVICE} />
