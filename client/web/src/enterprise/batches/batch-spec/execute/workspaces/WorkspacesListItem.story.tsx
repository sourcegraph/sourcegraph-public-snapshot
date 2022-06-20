import { useState } from 'react'

import { storiesOf } from '@storybook/react'

import { WebStory } from '../../../../../components/WebStory'
import { BatchSpecWorkspaceState } from '../../../../../graphql-operations'
import { mockWorkspace } from '../../batch-spec.mock'

import { WorkspacesListItem } from './WorkspacesListItem'

const { add } = storiesOf('web/batches/batch-spec/execute/workspaces/WorkspacesList', module).addDecorator(story => (
    <div className="list-group d-flex flex-column w-100">{story()}</div>
))

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

add('WorkspacesListItem', () => {
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
})
