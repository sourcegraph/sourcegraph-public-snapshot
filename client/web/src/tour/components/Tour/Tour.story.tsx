import React from 'react'

import type { Meta, DecoratorFn } from '@storybook/react'

import { MockTemporarySettings } from '@sourcegraph/shared/src/settings/temporary/testUtils'
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

export const VisitorsWithCompletedSteps: React.FunctionComponent<React.PropsWithChildren<unknown>> = () => (
    <MockTemporarySettings
        settings={{
            'onboarding.quickStartTour': {
                TourStorybook: {
                    completedStepIds: [visitorsTasks[0].steps[0].id],
                },
            },
        }}
    >
        <Tour telemetryService={NOOP_TELEMETRY_SERVICE} id="TourStorybook" tasks={visitorsTasks} />
    </MockTemporarySettings>
)

export const VisitorsWithCompletedTask: React.FunctionComponent<React.PropsWithChildren<unknown>> = () => (
    <MockTemporarySettings
        settings={{
            'onboarding.quickStartTour': {
                TourStorybook: {
                    completedStepIds: [...visitorsTasks[0].steps.map(step => step.id)],
                },
            },
        }}
    >
        <Tour telemetryService={NOOP_TELEMETRY_SERVICE} id="TourStorybook" tasks={visitorsTasks} />
    </MockTemporarySettings>
)

export const AuthenticatedDefault: React.FunctionComponent<React.PropsWithChildren<unknown>> = () => (
    <Tour
        telemetryService={NOOP_TELEMETRY_SERVICE}
        id="TourStorybook"
        tasks={authenticatedTasks}
        variant="horizontal"
    />
)

export const AuthenticatedWithCompletedSteps: React.FunctionComponent<React.PropsWithChildren<unknown>> = () => (
    <MockTemporarySettings
        settings={{
            'onboarding.quickStartTour': {
                TourStorybook: {
                    completedStepIds: [authenticatedTasks[2].steps[0].id],
                },
            },
        }}
    >
        <Tour
            telemetryService={NOOP_TELEMETRY_SERVICE}
            id="TourStorybook"
            tasks={authenticatedTasks}
            variant="horizontal"
        />
    </MockTemporarySettings>
)

export const AuthenticatedWithCompletedTask: React.FunctionComponent<React.PropsWithChildren<unknown>> = () => (
    <MockTemporarySettings
        settings={{
            'onboarding.quickStartTour': {
                TourStorybook: {
                    completedStepIds: [...authenticatedTasks[0].steps.map(step => step.id)],
                },
            },
        }}
    >
        <Tour
            telemetryService={NOOP_TELEMETRY_SERVICE}
            id="TourStorybook"
            tasks={authenticatedTasks}
            variant="horizontal"
        />
    </MockTemporarySettings>
)
