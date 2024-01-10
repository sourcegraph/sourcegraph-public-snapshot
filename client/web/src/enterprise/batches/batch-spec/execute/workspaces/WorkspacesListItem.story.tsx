import type { Decorator, Meta, StoryFn } from '@storybook/react'

import { WebStory } from '../../../../../components/WebStory'
import { BatchSpecWorkspaceState } from '../../../../../graphql-operations'
import { mockWorkspace } from '../../batch-spec.mock'

import { WorkspacesListItem } from './WorkspacesListItem'

const decorator: Decorator = story => <div className="list-group d-flex flex-column w-100">{story()}</div>

const config: Meta = {
    title: 'web/batches/batch-spec/execute/workspaces',
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

export const WorkspacesListItemStory: StoryFn = () => (
    <WebStory>
        {props => (
            <>
                {WORKSPACE_STATES.map(([key, state, isCached], index) => (
                    <>
                        {/* Not selected */}
                        <WorkspacesListItem
                            {...props}
                            key={key}
                            node={mockWorkspace(1, { state, cachedResultFound: isCached, id: index.toString() })}
                            executionURL="/fake/execution/url"
                        />
                        {/* Selected */}
                        <WorkspacesListItem
                            {...props}
                            key={key}
                            selectedNode={(100 + index).toString()}
                            node={mockWorkspace(1, {
                                state,
                                cachedResultFound: isCached,
                                id: (100 + index).toString(),
                            })}
                            executionURL="/fake/execution/url"
                        />
                    </>
                ))}
            </>
        )}
    </WebStory>
)

WorkspacesListItemStory.storyName = 'WorkspacesListItem'
