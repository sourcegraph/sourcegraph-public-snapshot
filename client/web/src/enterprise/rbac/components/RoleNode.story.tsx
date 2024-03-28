import type { Decorator, Meta, StoryFn } from '@storybook/react'
import { noop } from 'lodash'
import { MATCH_ANY_PARAMETERS, WildcardMockLink } from 'wildcard-mock-link'

import { getDocumentNode } from '@sourcegraph/http-client'
import { noOpTelemetryRecorder } from '@sourcegraph/shared/src/telemetry'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'

import { WebStory } from '../../../components/WebStory'
import { DELETE_ROLE, SET_PERMISSIONS } from '../backend'
import { mockRoles, mockPermissionsMap } from '../mock'

import { RoleNode } from './RoleNode'

const decorator: Decorator = story => <div className="p-3 container list-unstyled">{story()}</div>

const config: Meta = {
    title: 'web/src/site-admin/rbac/RoleNode',
    decorators: [decorator],
}

export default config

const mocks = new WildcardMockLink([
    {
        request: {
            query: getDocumentNode(DELETE_ROLE),
            variables: MATCH_ANY_PARAMETERS,
        },
        result: { data: { deleteRole: { alwaysNil: null } } },
        nMatches: Number.POSITIVE_INFINITY,
    },
    {
        request: {
            query: getDocumentNode(SET_PERMISSIONS),
            variables: MATCH_ANY_PARAMETERS,
        },
        result: { data: { setPermissions: { alwaysNil: null } } },
        nMatches: Number.POSITIVE_INFINITY,
    },
])

const [systemRole, nonSystemRole] = mockRoles.roles.nodes

export const SystemRole: StoryFn = () => (
    <WebStory>
        {() => (
            <MockedTestProvider link={mocks}>
                <RoleNode
                    allPermissions={mockPermissionsMap}
                    node={systemRole}
                    refetch={noop}
                    telemetryRecorder={noOpTelemetryRecorder}
                />
            </MockedTestProvider>
        )}
    </WebStory>
)

SystemRole.storyName = 'System role'

export const NonSystemRole: StoryFn = () => (
    <WebStory>
        {() => (
            <MockedTestProvider link={mocks}>
                <RoleNode
                    node={nonSystemRole}
                    refetch={noop}
                    allPermissions={mockPermissionsMap}
                    telemetryRecorder={noOpTelemetryRecorder}
                />
            </MockedTestProvider>
        )}
    </WebStory>
)

NonSystemRole.storyName = 'Non-system role'
