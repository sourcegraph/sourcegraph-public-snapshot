import React from 'react'

import type { Meta, Decorator } from '@storybook/react'

import { MockTemporarySettings } from '@sourcegraph/shared/src/settings/temporary/testUtils'
import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { WebStory } from '../../../components/WebStory'
import { authenticatedTasks } from '../../data'

import { Tour } from './Tour'

const decorator: Decorator = story => <WebStory>{() => <div className="container mt-3">{story()}</div>}</WebStory>

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

const userInfo = {
    repo: 'exampl/repo',
    email: 'user@example.com',
    language: 'TypeScript',
}

export default config

export const AuthenticatedDefault: React.FunctionComponent<React.PropsWithChildren<unknown>> = () => (
    <Tour
        telemetryService={NOOP_TELEMETRY_SERVICE}
        id="TourStorybook"
        tasks={authenticatedTasks}
        variant="horizontal"
        userInfo={userInfo}
        defaultSnippets={{}}
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
            userInfo={userInfo}
            defaultSnippets={{}}
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
            userInfo={userInfo}
            defaultSnippets={{}}
        />
    </MockTemporarySettings>
)
