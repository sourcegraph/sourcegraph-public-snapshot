import { ComponentMeta, ComponentStoryObj } from '@storybook/react'

import { Login } from './Login'
import { VSCodeStoryDecorator } from './storybook/VSCodeStoryDecorator'
import { VSCodeWrapper } from './utils/VSCodeApi'

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

const vscodeAPI: VSCodeWrapper = {
    postMessage: () => {},
    onMessage: () => () => {},
}

export default meta

export const Simple: ComponentStoryObj<typeof Login> = {
    render: () => (
        <div style={{ background: 'rgb(28, 33, 40)' }}>
            <Login onLogin={() => {}} isValidLogin={true} isAppInstalled={false} vscodeAPI={vscodeAPI} />
        </div>
    ),
}

export const InvalidLogin: ComponentStoryObj<typeof Login> = {
    render: () => (
        <div style={{ background: 'rgb(28, 33, 40)' }}>
            <Login onLogin={() => {}} isValidLogin={false} isAppInstalled={false} vscodeAPI={vscodeAPI} />
        </div>
    ),
}

export const LoginWithAppInstalled: ComponentStoryObj<typeof Login> = {
    render: () => (
        <div style={{ background: 'rgb(28, 33, 40)' }}>
            <Login onLogin={() => {}} isValidLogin={true} isAppInstalled={true} vscodeAPI={vscodeAPI} />
        </div>
    ),
}
