import type { Decorator, Meta, StoryFn } from '@storybook/react'
import { noop } from 'lodash'
import { MATCH_ANY_PARAMETERS, WildcardMockLink } from 'wildcard-mock-link'

import { getDocumentNode } from '@sourcegraph/http-client'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'

import { WebStory } from '../../../components/WebStory'
import { CREATE_ROLE } from '../backend'
import { mockPermissionsMap, mockRoles } from '../mock'

import { CreateRoleModal } from './CreateRoleModal'

const decorator: Decorator = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/src/site-admin/rbac',
    decorators: [decorator],
}

export default config

const [sampleRole] = mockRoles.roles.nodes

const mocks = new WildcardMockLink([
    {
        request: {
            query: getDocumentNode(CREATE_ROLE),
            variables: MATCH_ANY_PARAMETERS,
        },
        result: { data: { createRole: sampleRole } },
        nMatches: Number.POSITIVE_INFINITY,
    },
])

export const CreateRoleModalStory: StoryFn = () => (
    <WebStory>
        {() => (
            <MockedTestProvider link={mocks}>
                <CreateRoleModal onCancel={noop} afterCreate={noop} allPermissions={mockPermissionsMap} />
            </MockedTestProvider>
        )}
    </WebStory>
)
