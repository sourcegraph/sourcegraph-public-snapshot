import { DecoratorFn, Meta, Story } from '@storybook/react'
import { noop } from 'lodash'
import { MATCH_ANY_PARAMETERS, WildcardMockLink } from 'wildcard-mock-link'

import { getDocumentNode } from '@sourcegraph/http-client'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'

import { WebStory } from '../../../../components/WebStory'
import { GET_ALL_ROLES_AND_USER_ROLES, SET_ROLES_FOR_USER } from '../backend'

import { RoleAssignmentModal } from './RoleAssignmentModal'

const decorator: DecoratorFn = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/src/enterprise/site-admin/rbac',
    decorators: [decorator],
}

export default config

const mocks = new WildcardMockLink([
    {
        request: {
            query: getDocumentNode(SET_ROLES_FOR_USER),
            variables: MATCH_ANY_PARAMETERS,
        },
        result: { data: { setRoles: { alwaysNil: true } } },
        nMatches: Number.POSITIVE_INFINITY
    },
    {
        request: {
            query: getDocumentNode(GET_ALL_ROLES_AND_USER_ROLES),
            variables: MATCH_ANY_PARAMETERS,
        },
        result: {
            data: {
                roles: {
                    nodes: [
                        {
                            id: 'role-1',
                            name: 'USER',
                            system: true
                        },
                        {
                            id: 'role-2',
                            name: 'OPERATOR',
                            system: false
                        },
                        {
                            id: 'role-3',
                            name: 'TECH OPRS',
                            system: false
                        },
                    ]
                },
                node: {
                    __typename: 'User',
                    roles: {
                        nodes: [
                            {
                                id: 'role-1',
                                name: 'USER',
                                system: true
                            },
                            {
                                id: 'role-2',
                                name: 'OPERATOR',
                                system: false
                            },
                        ]
                    }
                }
            }
        },
        nMatches: Number.POSITIVE_INFINITY
    },
])

export const RoleAssignmentModalStory: Story = () => (
    <WebStory>
        {() => (
            <MockedTestProvider link={mocks}>
                <RoleAssignmentModal onCancel={noop} onSuccess={noop} />
            </MockedTestProvider>
        )}
    </WebStory>
)
