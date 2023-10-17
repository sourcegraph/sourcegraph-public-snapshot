import type { Decorator, Meta, StoryFn } from '@storybook/react'

import { BatchSpecWorkspaceState } from '@sourcegraph/shared/src/graphql-operations'

import { WebStory } from '../../../../../components/WebStory'

import { WorkspaceStateIcon } from './WorkspaceStateIcon'

const decorator: Decorator = story => <div className="p-3">{story()}</div>

const config: Meta = {
    title: 'web/batches/batch-spec/execute/workspaces/WorkspaceStateIcon',
    decorators: [decorator],
    argTypes: {
        cachedResultFound: {
            name: 'Cache Found',
            control: { type: 'boolean' },
        },
    },
    args: {
        cachedResultFound: false,
    },
}

export default config

export const WorkspaceStateIconStory: StoryFn = args => (
    <WebStory>
        {props => (
            <>
                {Object.values(BatchSpecWorkspaceState)
                    .sort()
                    .map(value => (
                        <div key={value} className="p-1">
                            {value}:{' '}
                            <WorkspaceStateIcon state={value} cachedResultFound={args.cachedResultFound} {...props} />
                        </div>
                    ))}
            </>
        )}
    </WebStory>
)

WorkspaceStateIconStory.storyName = 'WorkspaceStateIcon'
