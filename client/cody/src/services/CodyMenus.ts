import * as vscode from 'vscode'

import { isLocalApp, isDotCom } from '../chat/protocol'

export interface LoginMenuItem {
    id: string
    label: string
    description: string
    totalSteps: number
    uri: string
}

export interface LoginInput {
    endpoint: string | null | undefined
    token: string | null | undefined
}

export type AuthMenuType = 'signin' | 'signout' | 'switch'

function getItemLabel(uri: string, current: boolean): string {
    const icon = current ? '$(check) ' : ''
    if (isLocalApp(uri)) {
        return `${icon}Cody App`
    }
    if (isDotCom(uri)) {
        return `${icon}Sourcegraph.com`
    }
    return `${icon}${uri}`
}

export const AuthMenu = async (type: AuthMenuType, historyItems: string[]): Promise<LoginMenuItem | null> => {
    // Create option items
    const historySize = historyItems?.length
    const history =
        historySize > 0
            ? historyItems
                  ?.map((uri, i) => ({
                      id: uri,
                      label: getItemLabel(uri, type === 'switch' && i === historySize - 1),
                      description: type === 'signout' && i === 0 ? 'current' : '',
                      uri,
                  }))
                  .reverse()
            : []
    const seperator = [{ label: type === 'signin' ? 'previously used' : 'current', kind: -1 }]
    const optionItems = type === 'signout' ? history : [...LoginMenuOptionItems, ...seperator, ...history]
    const option = (await vscode.window.showQuickPick(optionItems, AuthMenuOptions[type])) as LoginMenuItem
    return option
}

// step 1 is to get the endpoint, step 2 is to get the token
export async function LoginStepInputBox(title: string, step: number, needToken: boolean): Promise<LoginInput | null> {
    // Get endpoint
    const options = LoginStepOptions[step - 1]
    options.title = title
    const endpoint = await vscode.window.showInputBox(options)
    if (!needToken || !endpoint) {
        return { endpoint, token: null }
    }
    return TokenInputBox(endpoint)
}

export async function TokenInputBox(endpoint: string): Promise<LoginInput | null> {
    // Get endpoint
    const options = LoginStepOptions[1]
    options.title = endpoint
    const token = await vscode.window.showInputBox(LoginStepOptions[1])
    return { endpoint, token }
}

export const AuthMenuOptions = {
    signin: {
        title: 'Other Sign in Options',
        placeholder: 'Select a sign in option',
        ignoreFocusOut: true,
    },
    signout: {
        title: 'Sign Out',
        placeHolder: 'Select instance to sign out',
        ignoreFocusOut: true,
    },
    switch: {
        title: 'Switch Account',
        placeHolder: 'Press Esc to cancel',
        ignoreFocusOut: true,
    },
}

export const LoginMenuOptionItems = [
    {
        id: 'token',
        label: 'Sign in with URL and Access Token',
        totalSteps: 2,
    },
    {
        id: 'dotcom',
        label: 'Sign in to Sourcegraph.com',
        totalSteps: 0,
    },
    {
        id: 'enterprise',
        label: 'Sign in to a Sourcegraph Enterprise Instance',
        description: 'Instance must be v5.1 or higher',
        totalSteps: 1,
        picked: true,
    },
]

const LoginStepOptions = [
    {
        prompt: 'Enter the URL to your Sourcegraph instance',
        placeholder: 'https://sourcegraph.mycompany.com/',
        password: false,
        ignoreFocusOut: true,
        totalSteps: 2,
        title: '',
        step: 1,
    },
    {
        prompt: 'Access Token',
        placeholder: 'Access Token',
        password: true,
        ignoreFocusOut: true,
        totalSteps: 2,
        title: 'Sign in with URL and Access Token',
        step: 2,
    },
]
