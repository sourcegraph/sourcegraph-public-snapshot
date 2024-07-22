import type { Meta, StoryFn } from '@storybook/react'

import { noOpTelemetryRecorder } from '@sourcegraph/shared/src/telemetry'

import { WebStory } from '../../components/WebStory'

import { NewTeamPage } from './NewTeamPage'

const config: Meta = {
    title: 'web/teams/NewTeamPage',
    parameters: {},
}
export default config

export const Default: StoryFn = function Default() {
    return <WebStory>{() => <NewTeamPage telemetryRecorder={noOpTelemetryRecorder} />}</WebStory>
}
