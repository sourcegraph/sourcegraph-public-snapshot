import * as vscode from 'vscode'

export interface LoginMenuItem {
    id: string
    label: string
    description: string
    totalSteps: number
}

export interface LoginInput {
    endpoint: string | null | undefined
    token: string | null | undefined
}

export const LoginMenuQuickPick = async (historyItems: string[]): Promise<LoginMenuItem | null> => {
    // Create option items
    const history =
        historyItems?.length > 0 ? historyItems?.map(endpoint => ({ id: endpoint, label: endpoint })).reverse() : []
    const seperator = [{ label: 'Last Signed in...', kind: -1 }]
    const optionItems = [...LoginMenuOptionItems, ...seperator, ...history]
    const option = (await vscode.window.showQuickPick(optionItems, LoginMenuOptions)) as LoginMenuItem
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
    const token = await vscode.window.showInputBox(LoginStepOptions[1])
    return { endpoint, token }
}

export const LoginMenuOptions = {
    title: 'Other Login Options',
    placeholder: 'Select a login option',
    ignoreFocusOut: true,
}

export const LoginMenuOptionItems = [
    {
        id: 'enterprise',
        label: 'Login to a Sourcegraph Enterprise Instance',
        description: 'Login to a Sourcegraph Enterprise Instance',
        totalSteps: 1,
    },
    {
        id: 'dotcom',
        label: 'Login to Sourcegraph.com',
        description: 'Login to Sourcegraph.com',
        totalSteps: 0,
    },
    {
        id: 'token',
        label: 'Login with URL and Access Token',
        description: 'Login with URL and Access Token',
        totalSteps: 2,
    },
]

const LoginStepOptions = [
    {
        prompt: 'Enter your Sourcegraph instance URL',
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
        title: 'Login with URL and Access Token',
        step: 2,
    },
]
