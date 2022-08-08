import { boolean, withKnobs } from '@storybook/addon-knobs'
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
            <>
                {Object.values(BatchSpecWorkspaceState)
                    .sort()
                    .map(value => (
                        <div key={value} className="p-1">
                            {value}:{' '}
                            <WorkspaceStateIcon
                                state={value}
                                cachedResultFound={boolean('Cache Found', false)}
                                {...props}
                            />
                        </div>
                    ))}
            </>
        )}
    </WebStory>
)

WorkspaceStateIconStory.storyName = 'WorkspaceStateIcon'
