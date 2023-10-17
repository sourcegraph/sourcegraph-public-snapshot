import React from 'react'

import type { Decorator, Meta, StoryFn } from '@storybook/react'
import { noop } from 'lodash'

import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'

import { WebStory } from '../../../components/WebStory'
import { mockRoles } from '../mock'

import { ConfirmDeleteRoleModal } from './ConfirmDeleteRoleModal'

const decorator: Decorator = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/src/site-admin/rbac',
    decorators: [decorator],
}

export default config

const mockOnConfirm = (event: React.FormEvent) => {
    event.preventDefault()
}
const batchChangeAdminRole = mockRoles.roles.nodes[1]

export const ConfirmDeleteRoleModalStory: StoryFn = () => (
    <WebStory>
        {() => (
            <MockedTestProvider>
                <ConfirmDeleteRoleModal role={batchChangeAdminRole} onCancel={noop} onConfirm={mockOnConfirm} />
            </MockedTestProvider>
        )}
    </WebStory>
)

ConfirmDeleteRoleModalStory.storyName = 'Confirm Delete role modal'
