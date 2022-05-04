import React from 'react'

import { Meta, DecoratorFn } from '@storybook/react'

import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { WebStory } from '../../../components/WebStory'
import { authenticatedTasks, visitorsTasks } from '../../data'

import { Tour } from './Tour'

const decorator: DecoratorFn = story => <WebStory>{() => <div className="container mt-3">{story()}</div>}</WebStory>

const config: Meta = {
    title: 'web/GettingStartedTour/Tour',
    decorators: [decorator],
    parameters: {
        component: Tour,
        chromatic: {
            enableDarkMode: true,
            disableSnapshot: false,
        },
    },
}

export default config

export const VisitorsDefault: React.FunctionComponent<React.PropsWithChildren<unknown>> = () => (
    <Tour telemetryService={NOOP_TELEMETRY_SERVICE} id="TourStorybook" tasks={visitorsTasks} />
)

export const AuthenticatedDefault: React.FunctionComponent<React.PropsWithChildren<unknown>> = () => (
    <Tour
        telemetryService={NOOP_TELEMETRY_SERVICE}
        id="TourStorybook"
        tasks={authenticatedTasks}
        variant="horizontal"
    />
)
