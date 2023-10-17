import type { Decorator, Meta, StoryFn } from '@storybook/react'
import { noop } from 'lodash'
import { MATCH_ANY_PARAMETERS, WildcardMockLink } from 'wildcard-mock-link'

import { getDocumentNode } from '@sourcegraph/http-client'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'

import { WebStory } from '../../../../components/WebStory'
import type { RoleFields } from '../../../../graphql-operations'
import { mockPermissions } from '../../../rbac/mock'
import { GET_ALL_ROLES_AND_USER_ROLES, SET_ROLES_FOR_USER } from '../backend'

import { RoleAssignmentModal } from './RoleAssignmentModal'

const decorator: Decorator = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/src/enterprise/site-admin/rbac',
    decorators: [decorator],
}

export default config

const nodes = mockPermissions.permissions.nodes

const buildMockRole = (id: number): RoleFields => ({
    __typename: 'Role',
    id: `role-${id}`,
    name: `Role ${id}`,
    system: false,
    permissions: {
        // A semi-random selection of permissions for each role.
        nodes:
            id % 5 === 0
                ? nodes
                : id % 4 === 0
                ? nodes.slice(0, 1)
                : id % 3 === 0
                ? nodes.slice(1, 3)
                : id % 2 === 0
                ? nodes.slice(2, 3)
                : nodes.slice(3, 4),
    },
})

const MOCK_SYSTEM_ROLE: RoleFields = {
    __typename: 'Role',
    id: 'role-1',
    name: 'USER',
    system: true,
    permissions: {
        nodes: [],
    },
}

const mocks = new WildcardMockLink([
    {
        request: {
            query: getDocumentNode(SET_ROLES_FOR_USER),
            variables: MATCH_ANY_PARAMETERS,
        },
        result: { data: { setRoles: { alwaysNil: true } } },
        nMatches: Number.POSITIVE_INFINITY,
    },
    {
        request: {
            query: getDocumentNode(GET_ALL_ROLES_AND_USER_ROLES),
            variables: MATCH_ANY_PARAMETERS,
        },
        result: {
            data: {
                roles: {
                    nodes: [MOCK_SYSTEM_ROLE, ...new Array(15).fill(0).map((_item, index) => buildMockRole(index + 2))],
                },
                node: {
                    __typename: 'User',
                    roles: {
                        nodes: [MOCK_SYSTEM_ROLE, buildMockRole(5)],
                    },
                },
            },
        },
        nMatches: Number.POSITIVE_INFINITY,
    },
])

export const RoleAssignmentModalStory: StoryFn = () => (
    <WebStory>
        {() => (
            <MockedTestProvider link={mocks}>
                <RoleAssignmentModal onCancel={noop} onSuccess={noop} user={{ id: 'user-1', username: 'user-1' }} />
            </MockedTestProvider>
        )}
    </WebStory>
)
