import { DecoratorFn, Meta, Story } from '@storybook/react'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'

import { WebStory } from '../../../components/WebStory'

import { PermissionList } from './Permissions'
import { mockPermissionsMap, mockRoles } from '../mock'

const decorator: DecoratorFn = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/src/site-admin/rbac/Permissions',
    decorators: [decorator],
}

export default config

const [systemRole, _, roleWithNoPermission] = mockRoles.roles.nodes

export const NoPermissions: Story = () => (
    <WebStory>
        {() => (
            <MockedTestProvider>
                <PermissionList role={roleWithNoPermission} allPermissions={mockPermissionsMap} />
            </MockedTestProvider>
        )}
    </WebStory>
)

NoPermissions.storyName = 'No permissions assigned'

export const AllPermissionsAssigned: Story = () => (
    <WebStory>
        {() => (
            <MockedTestProvider>
                <PermissionList role={systemRole} allPermissions={mockPermissionsMap} />
            </MockedTestProvider>
        )}
    </WebStory>
)

AllPermissionsAssigned.storyName = 'All permissions assigned'
