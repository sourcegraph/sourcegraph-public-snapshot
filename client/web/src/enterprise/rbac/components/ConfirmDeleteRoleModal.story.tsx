import React from 'react'

import { DecoratorFn, Meta, Story } from '@storybook/react'
import { noop } from 'lodash'
import { ApolloError } from '@apollo/client'

import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'

import { WebStory } from '../../../components/WebStory'

import { mockRoles } from '../mock'
import { ConfirmDeleteRoleModal } from './ConfirmDeleteRoleModal'

const decorator: DecoratorFn = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/src/site-admin/rbac/ConfirmDeleteRoleModal',
    decorators: [decorator],
}

export default config

const mockOnConfirm = (event: React.FormEvent) => {
    event.preventDefault()
}
const [_, batchChangeAdminRole] = mockRoles.roles.nodes

export const ConfirmDeleteRoleModalStory: Story = () => (
    <WebStory>
        {() => (
            <MockedTestProvider>
                <ConfirmDeleteRoleModal loading={false} error={undefined} role={batchChangeAdminRole} onCancel={noop} onConfirm={mockOnConfirm} />
            </MockedTestProvider>
        )}
    </WebStory>
)

ConfirmDeleteRoleModalStory.storyName = 'Confirm Delete role modal'

export const ConfirmDeleteRoleModalStoryLoading: Story = () => (
    <WebStory>
        {() => (
            <MockedTestProvider>
                <ConfirmDeleteRoleModal loading={true} error={undefined} role={batchChangeAdminRole} onCancel={noop} onConfirm={mockOnConfirm} />
            </MockedTestProvider>
        )}
    </WebStory>
)

ConfirmDeleteRoleModalStoryLoading.storyName = 'Confirm Delete role modal (loading)'

export const ConfirmDeleteRoleModalStoryWithError: Story = () => {
    const error = new ApolloError({ errorMessage: 'an error occurred' })
    return (
        <WebStory>
            {() => (
                <MockedTestProvider>
                    <ConfirmDeleteRoleModal loading={false} error={error} role={batchChangeAdminRole} onCancel={noop} onConfirm={mockOnConfirm} />
                </MockedTestProvider>
            )}
        </WebStory>
    )
}

ConfirmDeleteRoleModalStoryWithError.storyName = 'Confirm Delete role modal (with error)'
