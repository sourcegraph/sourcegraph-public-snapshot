import { boolean, select, withKnobs } from '@storybook/addon-knobs'
import { DecoratorFn, Meta, Story } from '@storybook/react'

import { BatchSpecWorkspaceState } from '@sourcegraph/shared/src/graphql-operations'

import { WebStory } from '../../../../../components/WebStory'

import { WorkspaceStateIcon } from './WorkspaceStateIcon'

const decorator: DecoratorFn = story => <div className="p-3">{story()}</div>

const config: Meta = {
    title: 'web/batches/batch-spec/execute/workspaces/WorkspaceStateIcon',
    decorators: [decorator, withKnobs],
}

export default config

export const WorkspaceStateIconStory: Story = () => (
    <WebStory>
        {props => (
            <WorkspaceStateIcon
                state={select(
                    'State',
                    [
                        BatchSpecWorkspaceState.PENDING,
                        BatchSpecWorkspaceState.QUEUED,
                        BatchSpecWorkspaceState.PROCESSING,
                        BatchSpecWorkspaceState.SKIPPED,
                        BatchSpecWorkspaceState.CANCELED,
                        BatchSpecWorkspaceState.CANCELING,
                        BatchSpecWorkspaceState.FAILED,
                        BatchSpecWorkspaceState.COMPLETED,
                    ],
                    BatchSpecWorkspaceState.COMPLETED
                )}
                cachedResultFound={boolean('Cache Found', false)}
                {...props}
            />
        )}
    </WebStory>
)

WorkspaceStateIconStory.storyName = 'WorkspaceStateIcon'
