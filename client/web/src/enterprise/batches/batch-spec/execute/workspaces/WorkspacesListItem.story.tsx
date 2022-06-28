import { useState } from 'react'

import { DecoratorFn, Meta, Story } from '@storybook/react'

import { WebStory } from '../../../../../components/WebStory'
import { BatchSpecWorkspaceState } from '../../../../../graphql-operations'
import { mockWorkspace } from '../../batch-spec.mock'

import { WorkspacesListItem } from './WorkspacesListItem'

const decorator: DecoratorFn = story => <div className="list-group d-flex flex-column w-100">{story()}</div>

const config: Meta = {
    title: 'web/batches/batch-spec/execute/workspaces/WorkspacesList',
    decorators: [decorator],
}

export default config

const WORKSPACE_STATES: [key: string, state: BatchSpecWorkspaceState, isCached: boolean][] = [
    ['pending', BatchSpecWorkspaceState.PENDING, false],
    ['queued', BatchSpecWorkspaceState.QUEUED, false],
    ['processing', BatchSpecWorkspaceState.PROCESSING, false],
    ['skipped', BatchSpecWorkspaceState.SKIPPED, false],
    ['canceled', BatchSpecWorkspaceState.CANCELED, false],
    ['canceling', BatchSpecWorkspaceState.CANCELING, false],
    ['failed', BatchSpecWorkspaceState.FAILED, false],
    ['completed', BatchSpecWorkspaceState.COMPLETED, false],
    ['completed-cached', BatchSpecWorkspaceState.COMPLETED, true],
]

export const WorkspacesListItemStory: Story = () => {
    const [selectedIndex, setSelectedIndex] = useState<number>()

    return (
        <WebStory>
            {props => (
                <>
                    {WORKSPACE_STATES.map(([key, state, isCached], index) => (
                        <WorkspacesListItem
                            {...props}
                            key={key}
                            isSelected={selectedIndex === index}
                            workspace={mockWorkspace(index, { state, cachedResultFound: isCached })}
                            onSelect={() => setSelectedIndex(index)}
                        />
                    ))}
                </>
            )}
        </WebStory>
    )
}

WorkspacesListItemStory.storyName = 'WorkspacesListItem'
