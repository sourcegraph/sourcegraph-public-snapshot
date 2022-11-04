import { DecoratorFn, Meta, Story } from '@storybook/react'
import { MATCH_ANY_PARAMETERS, WildcardMockLink } from 'wildcard-mock-link'

import { getDocumentNode } from '@sourcegraph/http-client'
import { MockedStoryProvider } from '@sourcegraph/storybook'

import { WebStory } from '../../../../../components/WebStory'
import { mockWorkspaces } from '../../batch-spec.mock'
import { BATCH_SPEC_WORKSPACES } from '../backend'

import { Workspaces } from './Workspaces'

const decorator: DecoratorFn = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/batches/batch-spec/execute/workspaces/Workspaces',
    decorators: [decorator],
}

export default config

const MOCKS = new WildcardMockLink([
    {
        request: {
            query: getDocumentNode(BATCH_SPEC_WORKSPACES),
            variables: MATCH_ANY_PARAMETERS,
        },
        result: { data: mockWorkspaces(50) },
        nMatches: Number.POSITIVE_INFINITY,
    },
])

export const WorkspacesStory: Story = () => (
    <WebStory>
        {props => (
            <MockedStoryProvider link={MOCKS}>
                <Workspaces batchSpecID="1" selectedNode="workspace1" executionURL="" {...props} />
            </MockedStoryProvider>
        )}
    </WebStory>
)

WorkspacesStory.storyName = 'Workspaces'
