import type { Meta } from '@storybook/react'

import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { WebStory } from '../../components/WebStory'

import { RepositoriesPopover } from './RepositoriesPopover'
import { MOCK_REQUESTS } from './RepositoriesPopover.mocks'

const Story: Meta = {
    title: 'web/RepositoriesPopover',

    decorators: [
        story => <WebStory mocks={MOCK_REQUESTS}>{() => <div className="container mt-3">{story()}</div>}</WebStory>,
    ],

    parameters: {
        component: RepositoriesPopover,
    },
}

export default Story

export const RepositoriesPopoverExample = () => (
    <RepositoriesPopover currentRepo="some-repo-id" telemetryService={NOOP_TELEMETRY_SERVICE} />
)
