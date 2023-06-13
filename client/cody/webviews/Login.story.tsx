import { ComponentMeta, ComponentStoryObj } from '@storybook/react'

import { AuthStatus, defaultAuthStatus, unauthenticatedStatus } from '../src/chat/protocol'

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
    ...defaultAuthStatus,
    authenticated: true,
    hasVerifiedEmail: true,
    requiresVerifiedEmail: false,
    siteHasCodyEnabled: true,
    siteVersion: '5.1.0',
}

const invalidAccessTokenAuthStatus: AuthStatus = { ...unauthenticatedStatus }

const requiresVerifiedEmailAuthStatus: AuthStatus = {
    ...defaultAuthStatus,
    authenticated: true,
    requiresVerifiedEmail: true,
    siteHasCodyEnabled: true,
    siteVersion: '5.1.0',
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
