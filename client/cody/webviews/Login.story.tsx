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
            enableDarkMode: true,
            disableSnapshot: false,
        },
    },
}

export default meta

export const Simple: ComponentStoryObj<typeof Login> = {
    render: () => <Login onLogin={() => {}} isValidLogin={true} />,
}

export const InvalidLogin: ComponentStoryObj<typeof Login> = {
    render: () => <Login onLogin={() => {}} isValidLogin={false} />,
}
