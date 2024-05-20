import type { Decorator, Meta, StoryFn } from '@storybook/react'
import { noop } from 'lodash'

import { PermissionNamespace } from '@sourcegraph/shared/src/graphql-operations'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'

import { WebStory } from '../../../components/WebStory'
import { PermissionsMap } from '../backend'
import { mockPermissionsMap, mockRoles } from '../mock'

import { PermissionsList } from './Permissions'

const decorator: Decorator = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/src/site-admin/rbac/Permissions',
    decorators: [decorator],
}

export default config

const [roleWithAllPermissions, roleWithOnePermission, roleWithNoPermission] = mockRoles.roles.nodes

const isChecked = (role: typeof roleWithAllPermissions): ((value: string) => boolean) => {
    const rolePermissions = role.permissions.nodes.reduce<Record<string, boolean>>((acc, node) => {
        acc[node.id] = true
        return acc
    }, {})
    return (value: string): boolean => rolePermissions[value]
}

const roleName = 'TEST-ROLE'

export const NoPermissions: StoryFn = () => (
    <WebStory>
        {() => (
            <MockedTestProvider>
                <PermissionsList
                    allPermissions={mockPermissionsMap}
                    onChange={noop}
                    onBlur={noop}
                    isChecked={isChecked(roleWithNoPermission)}
                    roleName={roleName}
                />
            </MockedTestProvider>
        )}
    </WebStory>
)

NoPermissions.storyName = 'No permissions assigned'

export const OnePermissionAssigned: StoryFn = () => (
    <WebStory>
        {() => (
            <MockedTestProvider>
                <PermissionsList
                    allPermissions={mockPermissionsMap}
                    onChange={noop}
                    onBlur={noop}
                    isChecked={isChecked(roleWithOnePermission)}
                    roleName={roleName}
                />
            </MockedTestProvider>
        )}
    </WebStory>
)

OnePermissionAssigned.storyName = 'One permission assigned'

export const AllPermissionsAssigned: StoryFn = () => (
    <WebStory>
        {() => (
            <MockedTestProvider>
                <PermissionsList
                    allPermissions={mockPermissionsMap}
                    onChange={noop}
                    onBlur={noop}
                    isChecked={isChecked(roleWithAllPermissions)}
                    roleName={roleName}
                />
            </MockedTestProvider>
        )}
    </WebStory>
)

AllPermissionsAssigned.storyName = 'All permissions assigned'

export const DotComPermissionIncluded: StoryFn = () => (
    <WebStory>
        {() => {
            const allPermissions: PermissionsMap = {
                ...mockPermissionsMap,
                [PermissionNamespace.PRODUCT_SUBSCRIPTIONS]: [
                    {
                        __typename: 'Permission',
                        id: 'test-03-01',
                        namespace: PermissionNamespace.PRODUCT_SUBSCRIPTIONS,
                        action: 'TEST',
                        displayName: 'PRODUCT_SUBSCRIPTIONS#TEST',
                    },
                ],
            }
            return (
                <MockedTestProvider>
                    <PermissionsList
                        allPermissions={allPermissions}
                        onChange={noop}
                        onBlur={noop}
                        isChecked={isChecked(roleWithAllPermissions)}
                        roleName={roleName}
                    />
                </MockedTestProvider>
            )
        }}
    </WebStory>
)
