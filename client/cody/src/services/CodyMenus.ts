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

export const QuickAuth = async (type: 'signin' | 'signout', historyItems: string[]): Promise<LoginMenuItem | null> => {
    // Create option items
    const isSignin = type === 'signin'
    const history =
        historyItems?.length > 0
            ? historyItems?.map((uri, i) => ({ id: uri, label: uri, description: i === 0 ? 'current' : '' })).reverse()
            : ([] as LoginMenuItem[])
    const seperator = [{ label: 'Last connected...', kind: -1 }]
    const optionItems = isSignin ? [...LoginMenuOptionItems, ...seperator, ...history] : history
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
        placeHolder: 'Select an account to sign out',
    },
}

export const LoginMenuOptionItems = [
    {
        id: 'enterprise',
        label: 'Sign in to a Sourcegraph Enterprise Instance',
        description: 'Sign in to a Sourcegraph Enterprise Instance',
        totalSteps: 1,
    },
    {
        id: 'dotcom',
        label: 'Sign in to Sourcegraph.com',
        description: 'Sign in to Sourcegraph.com',
        totalSteps: 0,
    },
    {
        id: 'token',
        label: 'Sign in with URL and Access Token',
        description: 'Sign in with URL and Access Token',
        totalSteps: 2,
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
