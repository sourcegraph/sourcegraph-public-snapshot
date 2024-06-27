import type { Decorator, Meta, StoryFn } from '@storybook/react'
import { of } from 'rxjs'

import { WebStory } from '../../../../../components/WebStory'
import { mockWorkspaces } from '../../batch-spec.mock'
import type { queryWorkspacesList as _queryWorkspacesList } from '../backend'

import { Workspaces } from './Workspaces'

const decorator: Decorator = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/batches/batch-spec/execute/workspaces/Workspaces',
    decorators: [decorator],
}

export default config

const queryWorkspacesList: typeof _queryWorkspacesList = () =>
    of(mockWorkspaces(50).node.workspaceResolution!.workspaces)

export const WorkspacesStory: StoryFn = () => (
    <WebStory>
        {props => (
            <Workspaces
                batchSpecID="1"
                selectedNode="workspace1"
                executionURL=""
                queryWorkspacesList={queryWorkspacesList}
                {...props}
            />
        )}
    </WebStory>
)

WorkspacesStory.storyName = 'Workspaces'
