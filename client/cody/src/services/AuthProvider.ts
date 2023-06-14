import * as vscode from 'vscode'

import { AuthStatus, DOTCOM_URL, LOCAL_APP_URL, isLoggedIn } from '../chat/protocol'
import { getAuthStatus } from '../chat/utils'

import { LocalStorage } from './LocalStorageProvider'
import { SecretStorage } from './SecretStorageProvider'

export class AuthProvider {
    private endpoint = DOTCOM_URL.href
    private endpointHistory: string[] = []
    private appScheme = vscode.env.uriScheme
    private authStatus: AuthStatus | null = null

    constructor(private secretStorage: SecretStorage, private localStorage: LocalStorage) {
        this.getEndpointHistory()
    }

    public async storeAuthInfo(endpoint: string, token: string): Promise<void> {
        this.setEndpoint(endpoint)
        await this.localStorage.saveEndpoint(endpoint)
        await this.secretStorage.storeToken(endpoint, token)
        this.getEndpointHistory()
    }

    public async makeAuthStatusFromCallback(uri: vscode.Uri, customHeaders: {}): Promise<AuthStatus | null> {
        const params = new URLSearchParams(uri.query)
        if (params.get('type') === 'app') {
            this.endpoint = LOCAL_APP_URL.href
        }
        const endpoint = this.endpoint
        const token = params.get('code')
        if (!token || !endpoint) {
            return null
        }
        await this.storeAuthInfo(endpoint, token)
        const authStatus = await getAuthStatus({
            serverEndpoint: endpoint,
            accessToken: token,
            customHeaders,
        })
        if (isLoggedIn(authStatus)) {
            void vscode.window.showInformationMessage('Token has been retrieved and updated successfully')
        }
        await vscode.commands.executeCommand('cody.chat.focus')
        return this.authStatus
    }

    public login(endpoint?: string): void {
        this.setEndpoint(endpoint)
        const quickPick = loginOptionsPicker(this.endpointHistory)
        quickPick.onDidChangeSelection(async selection => {
            quickPick.dispose()
            const title = selection[0].label
            switch (title) {
                case 'Login to a Sourcegraph Enterprise Instance':
                    this.loginWithURL(title, true)
                    break
                case 'Login to Sourcegraph.com':
                    this.redirectToEndpointLogin(true)
                    break
                case 'Login with URL and Access Token':
                    this.loginWithURL(title, false)
                    break
                default: {
                    // Auto log user if token for the selected instance was found in secret
                    const token = await this.secretStorage.get(title)
                    if (token) {
                        await this.storeAuthInfo(title, token)
                        return
                    }
                    this.loginWithToken(title, title)
                }
            }
        })
        quickPick.show()
    }

    public loginWithSourcegraph(): void {
        this.redirectToEndpointLogin(true)
    }

    private loginWithURL(title: string, redirect: boolean): void {
        const inputBox = inputStep(title, 1)
        inputBox.onDidAccept(() => {
            const endpoint = inputBox.value
            this.setEndpoint(endpoint)
            if (redirect) {
                this.redirectToEndpointLogin(false)
                return
            }
            this.loginWithToken(title, endpoint)
            inputBox.dispose()
        })
        inputBox.show()
    }

    private loginWithToken(title: string, endpoint: string): void {
        const inputBox = inputStep(title, 2)
        inputBox.onDidAccept(async () => {
            if (inputBox.value) {
                const token = inputBox.value
                await this.storeAuthInfo(endpoint, token)
            }
            inputBox.dispose()
        })
        inputBox.show()
    }

    private redirectToEndpointLogin(isDotCom: boolean): void {
        const endpoint = isDotCom ? DOTCOM_URL.href : this.endpoint
        if (!endpoint) {
            return
        }
        const authUri = new URL('/user/settings/tokens/new/callback', endpoint)
        authUri.searchParams.append('requestFrom', this.appScheme === 'vscode-insiders' ? 'CODY_INSIDERS' : 'CODY')
        this.setEndpoint(endpoint)
        // open external link
        void vscode.env.openExternal(vscode.Uri.parse(authUri.href))
    }

    private setEndpoint(endpoint?: string): void {
        if (!endpoint) {
            return
        }
        this.endpoint = endpoint
    }

    private getEndpointHistory(): void {
        this.endpointHistory = this.localStorage.getEndpointHistory() || []
    }
}

function loginOptionsPicker(historyItems: string[]): vscode.QuickPick<vscode.QuickPickItem> {
    const quickPick = vscode.window.createQuickPick()
    quickPick.title = 'Other Login Options'
    quickPick.placeholder = 'Select a login option '
    quickPick.ignoreFocusOut = true
    // Create options
    const options = loginOptions.map(item => ({ label: item.label }))
    const history = historyItems?.length > 0 ? historyItems?.map(endpoint => ({ label: endpoint })).reverse() : []
    const seperator = [{ label: 'Last Signed in...', kind: -1 }]
    quickPick.items = [...options, ...seperator, ...history]
    return quickPick
}

function inputStep(title: string, step: number): vscode.InputBox {
    const loginStep = loginStepOptions[step - 1]
    const inputBox = vscode.window.createInputBox()
    inputBox.title = title
    inputBox.step = step
    inputBox.totalSteps = 2
    inputBox.password = loginStep.prompt === 'Access Token'
    inputBox.prompt = loginStep.prompt
    inputBox.placeholder = loginStep.placeholder
    inputBox.ignoreFocusOut = true

    return inputBox
}

const loginOptions = [
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

const loginStepOptions = [
    {
        prompt: 'Enter your Sourcegraph instance URL',
        placeholder: 'http://sourcegraph.mycompany.com/',
    },
    {
        prompt: 'Access Token',
        placeholder: 'Access Token',
    },
]
