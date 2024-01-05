import type { Decorator, Meta, StoryFn } from '@storybook/react'
import { Subject } from 'rxjs'

import { ExternalServiceSyncJobState } from '../../graphql-operations'
import { WebStory } from '../WebStory'

import { ExternalServiceSyncJobNode } from './ExternalServiceSyncJobNode'

const decorator: Decorator = story => <WebStory>{() => <div className="p-3 container">{story()}</div>}</WebStory>

const config: Meta = {
    title: 'web/External services/ExternalServiceSyncJobNode',
    decorators: [decorator],
    parameters: {
        chromatic: { disableSnapshot: false },
    },
}

export default config

export const Default: StoryFn = () => (
    <div>
        {Object.values(ExternalServiceSyncJobState).map(state => (
            <ExternalServiceSyncJobNode
                key={state}
                onUpdate={new Subject()}
                node={{
                    __typename: 'ExternalServiceSyncJob',
                    id: '1',
                    state,
                    startedAt: '2023-08-25T06:31:00.000Z',
                    finishedAt: '2023-08-25T06:44:00.000Z',
                    failureMessage: null,
                    reposAdded: 12,
                    reposModified: 11,
                    reposUnmodified: 10,
                    reposSynced: 9,
                    reposDeleted: 8,
                    repoSyncErrors: 7,
                }}
            />
        ))}
    </div>
)
