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
    isLoggedIn as isAuthed,
    unauthenticatedStatus,
} from '../chat/protocol'
import { newAuthStatus } from '../chat/utils'
import { updateConfiguration } from '../configuration'
import { logEvent } from '../event-logger'
import { debug } from '../log'

import { AuthMenu, LoginStepInputBox, TokenInputBox } from './CodyMenus'
import { LocalStorage } from './LocalStorageProvider'
import { CODY_ACCESS_TOKEN_SECRET, SecretStorage } from './SecretStorageProvider'

// TODO (bee) log events
export class AuthProvider {
    private endpointHistory: string[] = []

    private appScheme = vscode.env.uriScheme
    private client: SourcegraphGraphQLAPIClient | null = null

    private authStatus: AuthStatus = unauthenticatedStatus

    constructor(
        private config: Pick<ConfigurationWithAccessToken, 'serverEndpoint' | 'accessToken' | 'customHeaders'>,
        private secretStorage: SecretStorage,
        private localStorage: LocalStorage
    ) {
        this.loadEndpointHistory()
        const lastEndpoint = localStorage?.getEndpoint()
        if (lastEndpoint) {
            this.autoSignin(lastEndpoint).catch(() => null)
        }
    }

    private async autoSignin(endpoint: string): Promise<void> {
        const token = await this.secretStorage.get(endpoint)
        await this.auth(endpoint, token || null)
    }

    public async signinMenu(): Promise<void> {
        const mode = this.authStatus.isLoggedIn ? 'switch' : 'signin'
        debug('AuthProvider:signinMenu', mode)
        logEvent('CodyVSCodeExtension:login:clicked')
        const item = await AuthMenu(mode, this.endpointHistory)
        if (!item) {
            return
        }
        switch (item?.id) {
            case 'enterprise': {
                const input = await LoginStepInputBox(item.uri, 1, false)
                if (!input?.endpoint) {
                    return
                }
                this.authStatus.endpoint = input.endpoint
                this.redirectToEndpointLogin(false)
                break
            }
            case 'dotcom':
                this.redirectToEndpointLogin(true)
                break
            case 'token': {
                const input = await LoginStepInputBox(item.uri, 1, true)
                if (!input?.endpoint || !input?.token) {
                    return
                }
                await this.auth(input.endpoint, input.token)
                break
            }
            default: {
                // Auto log user if token for the selected instance was found in secret
                const selectedEndpoint = item.uri
                const token = await this.secretStorage.get(selectedEndpoint)
                const authState = await this.auth(selectedEndpoint, token || null)
                if (!authState || authState?.isLoggedIn) {
                    return
                }
                const input = await TokenInputBox(item.uri)
                const isLoggedIn = (await this.auth(selectedEndpoint, input?.token || null))?.isLoggedIn
                if (isLoggedIn) {
                    void vscode.window.showInformationMessage(`Successfully signed in to ${selectedEndpoint}`)
                }
            }
        }
    }

    // Log user out of the current instance
    public async signoutMenu(): Promise<void> {
        logEvent('CodyVSCodeExtension:logout:clicked')
        const endpointQuickPickItem = this.authStatus.endpoint ? [this.authStatus.endpoint] : []
        const endpoint = await AuthMenu('signout', endpointQuickPickItem)
        if (!endpoint?.uri) {
            return
        }
        await this.signout(endpoint.uri)
        debug('AuthProvider:signoutMenu', endpoint.uri)
    }

    private async signout(endpoint: string): Promise<void> {
        await this.secretStorage.delete(CODY_ACCESS_TOKEN_SECRET)
        await updateConfiguration('serverEndpoint', '')
        await this.auth(endpoint, null)
    }

    public getAuthStatus(): AuthStatus {
        return this.authStatus
    }

    private async makeAuthStatus(
        config: Pick<ConfigurationWithAccessToken, 'serverEndpoint' | 'accessToken' | 'customHeaders'>
    ): Promise<AuthStatus> {
        const endpoint = config.serverEndpoint
        const token = config.accessToken
        if (!token || !endpoint) {
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

    // Verify and share auth status with chatview
    public async setAuthStatus(authStatus: AuthStatus): Promise<void> {
        if (this.authStatus === authStatus) {
            return
        }
        this.authStatus = authStatus
        await vscode.commands.executeCommand('cody.auth.sync', authStatus)
    }

    // It processes the authentication steps and stores the login info before sharing the auth status with chatview
    public async auth(
        endpoint: string,
        token: string | null,
        customHeaders = {}
    ): Promise<{ authStatus: AuthStatus; isLoggedIn: boolean } | null> {
        const config = {
            serverEndpoint: endpoint,
            accessToken: token,
            customHeaders: customHeaders || this.config.customHeaders,
        }
        const authStatus = await this.makeAuthStatus(config)
        const isLoggedIn = isAuthed(authStatus)
        authStatus.isLoggedIn = isLoggedIn
        await this.setAuthStatus(authStatus)
        await this.storeAuthInfo(endpoint, token)
        if (isLoggedIn) {
            void vscode.window.showInformationMessage(`You are signed in to ${endpoint}.`)
        }
        return { authStatus, isLoggedIn }
    }

    // Register URI Handler (vscode://sourcegraph.cody-ai) for:
    // - Deep linking into VS Code with Cody focused (e.g. from the App setup)
    // - Resolving token sending back from sourcegraph.com and App
    public async tokenCallbackHandler(uri: vscode.Uri, customHeaders: {}): Promise<void> {
        const params = new URLSearchParams(uri.query)
        const isApp = params.get('type') === 'app'
        const token = params.get('code')
        const endpoint = isApp ? LOCAL_APP_URL.href : this.authStatus.endpoint
        if (!token || !endpoint) {
            return
        }
        await this.storeAuthInfo(endpoint, token)
        const authState = await this.auth(endpoint, token, customHeaders)
        if (authState?.isLoggedIn) {
            const actionButtonLabel = 'Get Started'
            const successMessage = isApp ? 'Connected to Cody App' : 'Signed in to Sourcegraph.com'
            const action = await vscode.window.showInformationMessage(successMessage, actionButtonLabel)
            if (action === actionButtonLabel) {
                await vscode.commands.executeCommand('cody.chat.focus')
            }
        }
    }

    private redirectToEndpointLogin(isDotCom: boolean): void {
        const endpoint = isDotCom ? DOTCOM_URL.href : this.authStatus.endpoint
        if (!endpoint) {
            return
        }
        const authUri = new URL('/user/settings/tokens/new/callback', endpoint)
        authUri.searchParams.append('requestFrom', this.appScheme === 'vscode-insiders' ? 'CODY_INSIDERS' : 'CODY')
        this.authStatus.endpoint = endpoint
        // open external link
        void vscode.env.openExternal(vscode.Uri.parse(authUri.href))
    }

    private loadEndpointHistory(): void {
        this.endpointHistory = this.localStorage.getEndpointHistory() || []
    }

    private async storeAuthInfo(endpoint: string | null | undefined, token: string | null | undefined): Promise<void> {
        debug('AuthProvider:storeAuthInfo:init', endpoint || '')
        if (!endpoint) {
            return
        }
        await this.localStorage.saveEndpoint(endpoint)
        await updateConfiguration('serverEndpoint', endpoint)
        if (token) {
            await this.secretStorage.storeToken(endpoint, token)
        }
        this.loadEndpointHistory()
        debug('AuthProvider:storeAuthInfo:stored', endpoint || '')
    }
}
