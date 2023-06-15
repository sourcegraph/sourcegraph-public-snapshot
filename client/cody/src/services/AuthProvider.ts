import * as vscode from 'vscode'

import { ConfigurationWithAccessToken } from '@sourcegraph/cody-shared/src/configuration'
import { SourcegraphGraphQLAPIClient } from '@sourcegraph/cody-shared/src/sourcegraph-api/graphql'
import { isError } from '@sourcegraph/cody-shared/src/utils'

import {
    AuthStatus,
    DOTCOM_URL,
    LOCAL_APP_URL,
    defaultAuthStatus,
    isLocalApp,
    isLoggedIn,
    unauthenticatedStatus,
} from '../chat/protocol'
import { newAuthStatus } from '../chat/utils'

import { LoginMenuQuickPick, LoginStepInputBox } from './CodyMenus'
import { LocalStorage } from './LocalStorageProvider'
import { CODY_ACCESS_TOKEN_SECRET, SecretStorage } from './SecretStorageProvider'

export class AuthProvider {
    private endpoint: string | null = null
    private endpointHistory: string[] = []
    private appScheme = vscode.env.uriScheme
    private authStatus: AuthStatus | null = null
    private client: SourcegraphGraphQLAPIClient | null = null

    constructor(
        private config: Pick<ConfigurationWithAccessToken, 'serverEndpoint' | 'accessToken' | 'customHeaders'>,
        private secretStorage: SecretStorage,
        private localStorage: LocalStorage
    ) {
        this.loadEndpointHistory()
    }

    public async login(endpoint?: string): Promise<void> {
        this.setEndpoint(endpoint)
        const item = await LoginMenuQuickPick(this.endpointHistory)
        if (!item) {
            return
        }
        switch (item?.id) {
            case 'enterprise': {
                const input = await LoginStepInputBox(item.label, 1, false)
                if (!input?.endpoint) {
                    return
                }
                this.setEndpoint(input?.endpoint)
                this.redirectToEndpointLogin(false)
                break
            }
            case 'dotcom':
                this.redirectToEndpointLogin(true)
                break
            case 'token': {
                const input = await LoginStepInputBox(item.label, 1, true)
                if (!input?.endpoint || !input?.token) {
                    console.log('No endpoint or token provided')
                    return
                }
                await this.auth(input?.endpoint, input?.token || null)
                break
            }
            default: {
                // Auto log user if token for the selected instance was found in secret
                const selectedEndpoint = item.label
                this.setEndpoint(selectedEndpoint)
                const token = await this.secretStorage.get(selectedEndpoint)
                const authedUser = await this.auth(selectedEndpoint, token || null)
                if (authedUser) {
                    return
                }
                const input = await LoginStepInputBox(item.label, 2, false)
                await this.auth(selectedEndpoint, input?.token || null)
            }
        }
    }

    public async logout(): Promise<void> {
        await this.secretStorage.delete(CODY_ACCESS_TOKEN_SECRET)
        if (this.endpoint) {
            await this.secretStorage.delete(this.endpoint)
        }
        await vscode.commands.executeCommand('setContext', 'cody.activated', false)
        this.endpoint = null
        this.authStatus = null
    }

    public async getAuthStatus(
        config: Pick<ConfigurationWithAccessToken, 'serverEndpoint' | 'accessToken' | 'customHeaders'>
    ): Promise<AuthStatus> {
        const endpoint = config.serverEndpoint
        if (!config.accessToken || !endpoint) {
            return { ...defaultAuthStatus, endpoint }
        }
        // Cache the config and the GraphQL client
        if (this.config !== config || !this.client) {
            this.config = config
            this.client = new SourcegraphGraphQLAPIClient(config)
        }
        // Version is for frontend to check if Cody is not enabled due to unsupported version when siteHasCodyEnabled is false
        const { enabled, version } = await this.client.isCodyEnabled()
        const isDotComOrApp = this.client.isDotCom() || isLocalApp(endpoint)
        if (!isDotComOrApp) {
            const currentUserID = await this.client.getCurrentUserId()
            return newAuthStatus(endpoint, isDotComOrApp, !isError(currentUserID), false, enabled, version)
        }
        const userInfo = await this.client.getCurrentUserIdAndVerifiedEmail()
        return isError(userInfo)
            ? { ...unauthenticatedStatus, endpoint }
            : newAuthStatus(endpoint, isDotComOrApp, !!userInfo.id, userInfo.hasVerifiedEmail, true, version)
    }

    // It process the authetication steps and store the login info.
    // Returns Auth state
    public async auth(endpoint: string, token: string | null, customHeaders?: {}): Promise<boolean> {
        await this.storeAuthInfo(endpoint, token)
        const config = {
            serverEndpoint: endpoint,
            accessToken: token,
            customHeaders: customHeaders || this.config.customHeaders,
        }
        const authStatus = await this.getAuthStatus(config)
        await vscode.commands.executeCommand('cody.auth.verify', authStatus)
        return isLoggedIn(authStatus)
    }

    // For Uri Handler
    public async tokenCallbackHandler(uri: vscode.Uri, customHeaders: {}): Promise<AuthStatus | null> {
        const params = new URLSearchParams(uri.query)
        const isApp = params.get('type') === 'app'
        if (isApp) {
            this.endpoint = LOCAL_APP_URL.href
        }
        const endpoint = this.endpoint
        const token = params.get('code')
        if (!token || !endpoint) {
            return null
        }
        await this.storeAuthInfo(endpoint, token)
        const authStatus = await this.auth(endpoint, token, customHeaders)
        if (authStatus) {
            const actionButtonLabel = 'Get Started'
            const successMessage = isApp ? 'Connected to Cody App' : 'Logged in to sourcegraph.com'
            const action = await vscode.window.showInformationMessage(successMessage, actionButtonLabel)
            if (action === actionButtonLabel) {
                await vscode.commands.executeCommand('cody.chat.focus')
            }
        }
        return this.authStatus
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

    private loadEndpointHistory(): void {
        this.endpointHistory = this.localStorage.getEndpointHistory() || []
    }

    public async storeAuthInfo(endpoint: string | null | undefined, token: string | null | undefined): Promise<void> {
        if (!endpoint || !token) {
            return
        }
        this.setEndpoint(endpoint)
        await this.localStorage.saveEndpoint(endpoint)
        await this.secretStorage.storeToken(endpoint, token)
        this.loadEndpointHistory()
    }
}
