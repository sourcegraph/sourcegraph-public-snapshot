import * as vscode from 'vscode'

export interface LocalProcess {
    arch: string
    homeDir: string | undefined
    os: string
    isAppInstalled: boolean
}

const LOCAL_APP_LOCATIONS: { [key: string]: string[] } = {
    // Only apply silicon is supported
    darwin: [
        '/Applications/Sourcegraph.app',
        '/Applications/Cody.app',
        '/Library/Application Support/com.sourcegraph.cody',
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

    private isSupported = false
    private isInstalled = false
    private localAppMarkers: string[] | undefined

    private onChange: OnChangeCallback
    private _watchers: vscode.Disposable[] = []

    constructor(options: { onChange: OnChangeCallback }) {
        this.onChange = options.onChange
        this.platformName = process.platform
        this.arch = process.arch
        this.homeDir = process.env.HOME
        this.localAppMarkers = LOCAL_APP_LOCATIONS[this.platformName]
        // Only Mac Silicon (M1 chip) is supported
        this.isSupported = this.platformName === 'darwin' && this.arch === 'arm64'
        this.start()
    }

    public async detect(): Promise<void> {
        const startCondition = this.canStart()
        if (!startCondition || !this.localAppMarkers) {
            return
        }
        const foundPaths = new Set()
        for (const marker of this.localAppMarkers) {
            const markerExists = await pathExists(marker)
            if (markerExists) {
                foundPaths.add(marker)
            }
        }
        // Check if Sourcegraph.app/Cody.aoo AND com.sourcegraph.cody are found
        this.fire(foundPaths.size === 2)
    }

    private fire(state: boolean): void {
        if (!this.isSupported) {
            return
        }
        this.onChange(state)
        this.isInstalled = state
    }

    public start(): void {
        const markers = this.localAppMarkers
        const startCondition = this.canStart()
        if (!startCondition || !markers || !this.homeDir) {
            return
        }
        for (const marker of markers) {
            const watchPattern = new vscode.RelativePattern(this.homeDir, `${marker}`)
            const watcher = vscode.workspace.createFileSystemWatcher(watchPattern)
            watcher.onDidChange(() => this.detect())
            watcher.onDidCreate(() => this.detect())
            watcher.onDidDelete(() => this.detect())
            this._watchers.push(watcher)
        }
        void this.detect()
    }

    private canStart(): boolean {
        if (!this.isSupported || this._watchers.length || this.isInstalled || !this.homeDir) {
            return false
        }
        return true
    }

    public getProcessInfo(): LocalProcess {
        return {
            arch: this.arch,
            os: this.platformName,
            homeDir: this.homeDir,
            isAppInstalled: this.isInstalled,
        }
    }

    public get isLocalAppInstalled(): boolean {
        return this.isInstalled
    }

    public dispose(): void {
        if (!this._watchers.length) {
            return
        }
        for (const watcher of this._watchers) {
            watcher.dispose()
        }
        this._watchers = []
    }
}
