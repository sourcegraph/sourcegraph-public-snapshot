import * as vscode from 'vscode'

import { version } from '../../package.json'
import { LOCAL_APP_URL, LocalEnv } from '../chat/protocol'

export interface LocalAppPaths {
    [os: string]: {
        dir: string
        file: string
    }[]
}

const LOCAL_APP_LOCATIONS: LocalAppPaths = {
    darwin: [
        {
            dir: '/Applications/',
            file: 'Sourcegraph.app',
        },
        {
            dir: '/Applications/',
            file: 'Cody.app',
        },
        {
            dir: '~/Library/Application Support/com.sourcegraph.cody/',
            file: 'site.config.json',
        },
    ],
}

async function pathExists(path: string): Promise<boolean> {
    try {
        await vscode.workspace.fs.stat(vscode.Uri.file(path))
        return true
    } catch {
        return false
    }
}

type OnChangeCallback = (value: boolean) => void

/**
 * Detects whether the user has the Sourcegraph app installed locally.
 */
export class LocalAppDetector implements vscode.Disposable {
    private arch: string
    private platformName: string
    private homeDir: string | undefined

    // Check if the platform is supported and the user has a home directory
    private isSupported: boolean
    private isInstalled = false
    private isRunning = false
    private localAppMarkers
    private appFsPaths: string[] = []

    // TODO: remove this once the experimental period for connect app is over
    private isAppConnectEnabled = false

    private onChange: OnChangeCallback
    private _watchers: vscode.Disposable[] = []

    constructor(options: { onChange: OnChangeCallback }) {
        this.onChange = options.onChange
        this.platformName = process.platform
        this.arch = process.arch
        this.homeDir = process.env.HOME
        this.localAppMarkers = LOCAL_APP_LOCATIONS[this.platformName]
        // Only Mac is supported for now
        this.isSupported = this.platformName === 'darwin' && this.homeDir !== undefined
        const codyConfiguration = vscode.workspace.getConfiguration('cody')
        // TODO: remove this once the experimental period for connect app is over
        this.isAppConnectEnabled = codyConfiguration.get<boolean>('experimental.app.connect') ?? false
        this.init()
    }

    private init(): void {
        // if conditions are not met, this will be a noop
        if (this._watchers.length || !this.isSupported || !this.homeDir || this.isInstalled) {
            return
        }
        // Create filePaths and file watchers
        const markers = this.localAppMarkers
        for (const marker of markers) {
            const dirPath = this.expandHomeDir(marker.dir)
            const dirUri = vscode.Uri.file(dirPath)
            const watchPattern = new vscode.RelativePattern(dirUri, marker.file)
            const watcher = vscode.workspace.createFileSystemWatcher(watchPattern)
            watcher.onDidChange(() => this.detect())
            this._watchers.push(watcher)
            this.appFsPaths.push(dirPath + marker.file)
        }
        void this.detect()
    }

    private async detect(): Promise<void> {
        if (!this.isSupported || !this.appFsPaths.length) {
            return
        }
        if (!this.isInstalled) {
            if (await Promise.any(this.appFsPaths.map(file => pathExists(file)))) {
                this.isInstalled = true
                await this.fetch()
                this.fire()
            }
            return
        }
    }

    // Check if App is running
    public async fetch(): Promise<void> {
        if (!this.isInstalled || this.isRunning) {
            return
        }
        const response = await fetch(`${LOCAL_APP_URL.href}__version`)
        if (response.status === 200) {
            this.isRunning = true
            console.log('App is running.')
            return
        }
    }

    public getProcessInfo(): LocalEnv {
        void this.fetch()
        return {
            os: this.platformName,
            arch: this.arch,
            homeDir: this.homeDir,
            uriScheme: vscode.env.uriScheme,
            appName: vscode.env.appName,
            extensionVersion: version,
            isAppInstalled: this.isInstalled,
            isAppRunning: this.isRunning,
            isAppConnectEnabled: this.isAppConnectEnabled,
        }
    }

    private expandHomeDir(path: string): string {
        if (path.startsWith('~/')) {
            return path.replace('~', process.env.HOME || '')
        }
        return path
    }

    // We can dispose the file watcher when app is found or when user has logged in
    private fire(): void {
        console.info('app found')
        this.onChange(true)
        this.dispose()
    }

    public dispose(): void {
        for (const watcher of this._watchers) {
            watcher.dispose()
        }
        this._watchers = []
        this.appFsPaths = []
    }
}
