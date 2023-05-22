import { ComponentMeta, ComponentStoryObj } from '@storybook/react'

import { AuthStatus } from '../src/chat/protocol'

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

const validAuthStatus: AuthStatus = {
    authenticated: true,
    showInvalidAccessTokenError: false,
    hasVerifiedEmail: false,
    requiresVerifiedEmail: false,
}

const invalidAccessTokenAuthStatus: AuthStatus = {
    authenticated: false,
    showInvalidAccessTokenError: true,
    hasVerifiedEmail: false,
    requiresVerifiedEmail: false,
}

const requiresVerifiedEmailAuthStatus: AuthStatus = {
    authenticated: true,
    showInvalidAccessTokenError: false,
    hasVerifiedEmail: false,
    requiresVerifiedEmail: true,
}

export default meta

export const Simple: ComponentStoryObj<typeof Login> = {
    render: () => (
        <div style={{ background: 'rgb(28, 33, 40)' }}>
            <Login onLogin={() => {}} authStatus={validAuthStatus} isAppInstalled={false} vscodeAPI={vscodeAPI} />
        </div>
    ),
}

export const InvalidLogin: ComponentStoryObj<typeof Login> = {
    render: () => (
        <div style={{ background: 'rgb(28, 33, 40)' }}>
            <Login
                onLogin={() => {}}
                authStatus={invalidAccessTokenAuthStatus}
                isAppInstalled={false}
                vscodeAPI={vscodeAPI}
            />
        </div>
    ),
}

export const UnverifiedEmailLogin: ComponentStoryObj<typeof Login> = {
    render: () => (
        <div style={{ background: 'rgb(28, 33, 40)' }}>
            <Login
                onLogin={() => {}}
                authStatus={requiresVerifiedEmailAuthStatus}
                isAppInstalled={false}
                vscodeAPI={vscodeAPI}
            />
        </div>
    ),
}

export const LoginWithAppInstalled: ComponentStoryObj<typeof Login> = {
    render: () => (
        <div style={{ background: 'rgb(28, 33, 40)' }}>
            <Login onLogin={() => {}} authStatus={validAuthStatus} isAppInstalled={true} vscodeAPI={vscodeAPI} />
        </div>
    ),
}
