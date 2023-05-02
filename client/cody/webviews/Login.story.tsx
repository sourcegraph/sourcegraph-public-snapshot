import { ComponentMeta, ComponentStoryObj } from '@storybook/react'

import { Login } from './Login'
import { VSCodeStoryDecorator } from './storybook/VSCodeStoryDecorator'

const meta: ComponentMeta<typeof Login> = {
    title: 'cody/Login',
    component: Login,
    decorators: [VSCodeStoryDecorator],
    parameters: {
        component: Login,
        chromatic: {
            disableSnapshot: false,
        },
    },
}

export default meta

export const Simple: ComponentStoryObj<typeof Login> = {
    render: () => (
        <div style={{ background: 'rgb(28, 33, 40)' }}>
            <Login onLogin={() => {}} isValidLogin={true} />
        </div>
    ),
}

export const InvalidLogin: ComponentStoryObj<typeof Login> = {
    render: () => (
        <div style={{ background: 'rgb(28, 33, 40)' }}>
            <Login onLogin={() => {}} isValidLogin={false} />
        </div>
    ),
}
