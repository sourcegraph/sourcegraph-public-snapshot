import * as vscode from 'vscode'

import { version } from '../../package.json'
import { LOCAL_APP_URL, LocalEnv, isOsSupportedByApp } from '../chat/protocol'
import { debug } from '../log'

import { AppJson, LOCAL_APP_LOCATIONS } from './LocalAppFsPaths'
import { SecretStorage } from './SecretStorageProvider'

type OnChangeCallback = (type: string) => Promise<void>
/**
 * Detects whether the user has the Sourcegraph app installed locally.
 */
export class LocalAppDetector implements vscode.Disposable {
    private localEnv: LocalEnv

    // Check if the platform is supported and the user has a home directory
    private isSupported = false

    private localAppMarkers
    private appFsPaths: string[] = []
    private tokenFsPath: vscode.Uri | null = null

    private _watchers: vscode.Disposable[] = []
    private onChange: OnChangeCallback

    constructor(private secretStorage: SecretStorage, options: { onChange: OnChangeCallback }) {
        this.onChange = options.onChange
        this.localEnv = envInit
        this.localAppMarkers = LOCAL_APP_LOCATIONS[this.localEnv.os]
        // Only Mac is supported for now
        this.isSupported =
            isOsSupportedByApp(this.localEnv.os, this.localEnv.arch) && this.localEnv.homeDir !== undefined
    }

    public async getProcessInfo(isLoggedIn = false): Promise<LocalEnv> {
        if (isLoggedIn && this._watchers.length > 0) {
            this.dispose()
        }
        await this.fetchServer()
        return this.localEnv
    }

    public async init(): Promise<void> {
        this.dispose()
        debug('LocalAppDetector:init', 'initializing')
        const homeDir = this.localEnv.homeDir
        // if conditions are not met, this will be a noop
        if (!this.isSupported || !homeDir) {
            debug('LocalAppDetector:init:failed', 'osNotSupported')
            return
        }
        // Create filePaths and file watchers
        const markers = this.localAppMarkers
        for (const marker of markers) {
            const dirPath = expandHomeDir(marker.dir, homeDir)
            const dirUri = vscode.Uri.file(dirPath)
            const watchPattern = new vscode.RelativePattern(dirUri, marker.file)
            const watcher = vscode.workspace.createFileSystemWatcher(watchPattern)
            watcher.onDidChange(() => this.fetchApp())
            this._watchers.push(watcher)
            this.appFsPaths.push(dirPath + marker.file)
            if (marker.hasToken) {
                this.tokenFsPath = vscode.Uri.file(dirPath + marker.file)
            }
        }
        debug('LocalAppDetector:init', 'initialized')
        await this.fetchApp()
    }

    // Check if App is installed
    private async fetchApp(): Promise<void> {
        if (this.localEnv.isAppInstalled || !this.appFsPaths) {
            return
        }
        debug('LocalAppDetector:fetchApp', 'initializing')
        if (await Promise.any(this.appFsPaths.map(file => pathExists(file)))) {
            debug('LocalAppDetector:fetchApp', 'found')
            this.localEnv.isAppInstalled = true
            this.appFsPaths = []
            await this.found('app')
            await this.fetchToken()
            return
        }
        debug('LocalAppDetector:detect:fetchApp', 'failed')
    }

    // Get token from app.json if it exists
    private async fetchToken(): Promise<void> {
        if (!this.tokenFsPath || this.localEnv.hasAppJson) {
            return
        }
        const appJson = await loadAppJson(this.tokenFsPath)
        if (!appJson) {
            debug('LocalAppDetector:fetchToken:loadAppJson', 'failed')
            return
        }
        const token = appJson.token
        // Once the token is found, we can stop watching the files
        if (token?.length) {
            this.localEnv.hasAppJson = true
            this.tokenFsPath = null
            await this.found('token')
            await this.secretStorage.storeToken(LOCAL_APP_URL.href, token)
            await this.fetchServer()
        }
        debug('LocalAppDetector:fetchToken', 'found')
    }

    // Check if App is running
    private async fetchServer(): Promise<void> {
        if (this.localEnv.isAppRunning) {
            return
        }
        debug('LocalAppDetector:fetchServer', 'initializing')
        try {
            const response = await fetch(`${LOCAL_APP_URL.href}__version`)
            if (response.status === 200) {
                debug('LocalAppDetector:fetchServer', 'found')
                this.localEnv.isAppRunning = true
                await this.found('server')
            }
            if (!this.localEnv.hasAppJson) {
                await this.fetchToken()
            }
        } catch {
            debug('LocalAppDetector:fetchServer', 'failed')
            return
        }
    }

    // Notify the caller that the app has been found
    // NOTE: Call this function only when the app is found
    private async found(type: 'app' | 'token' | 'server'): Promise<void> {
        this.localEnv.isAppInstalled = true
        await this.onChange(type)
        debug('LocalAppDetector:found', type)
    }

    // We can dispose the file watcher when app is found or when user has logged in
    public dispose(): void {
        for (const watcher of this._watchers) {
            watcher.dispose()
        }
        this._watchers = []
        this.appFsPaths = []
        this.tokenFsPath = null
    }
}

// Utility functions
async function pathExists(path: string): Promise<boolean> {
    try {
        await vscode.workspace.fs.stat(vscode.Uri.file(path))
        return true
    } catch {
        return false
    }
}

function expandHomeDir(path: string, homeDir: string | null): string {
    if (homeDir && path.startsWith('~/')) {
        return path.replace('~', homeDir)
    }
    return path
}

async function loadAppJson(uri: vscode.Uri): Promise<AppJson | null> {
    try {
        const data = await vscode.workspace.fs.readFile(uri)
        return JSON.parse(data.toString()) as AppJson
    } catch {
        return null
    }
}

const envInit = {
    os: process.platform,
    arch: process.arch,
    homeDir: process.env.HOME,
    uriScheme: vscode.env.uriScheme,
    appName: vscode.env.appName,
    extensionVersion: version,
    isAppInstalled: false,
    isAppRunning: false,
    hasAppJson: false,
}
