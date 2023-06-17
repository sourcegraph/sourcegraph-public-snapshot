import * as vscode from 'vscode'

export interface LocalProcess {
    arch?: string
    homeDir?: string | undefined
    os?: string
    isAppInstalled: boolean
}

const LOCAL_APP_LOCATIONS: { [key: string]: string[] } = {
    // Only Apple silicon is supported
    darwin: [
        '~/Library/Application Support/com.sourcegraph.cody',
        '/Applications/Sourcegraph.app',
        '/Applications/Cody.app',
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

function expandHomeDir(path: string): string {
    if (path.startsWith('~/')) {
        return path.replace('~', process.env.HOME || '')
    }
    return path
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
    private localAppMarkers: string[]

    private onChange: OnChangeCallback
    private _watchers: vscode.Disposable[] = []

    constructor(options: { onChange: OnChangeCallback }) {
        this.onChange = options.onChange
        this.platformName = process.platform
        this.arch = process.arch
        this.homeDir = process.env.HOME
        this.localAppMarkers = LOCAL_APP_LOCATIONS[this.platformName] || []
        // Only Mac is supported for now
        this.isSupported = this.platformName === 'darwin' && this.homeDir !== undefined
        this.start()
    }

    public async detect(): Promise<void> {
        if (!this.isSupported) {
            return
        }
        if (await Promise.any(this.localAppMarkers.map(marker => pathExists(expandHomeDir(marker))))) {
            this.isInstalled = true
        }
        this.fire()
    }

    public start(): void {
        if (this._watchers.length || !this.isSupported || !this.homeDir || this.isInstalled) {
            return
        }
        void this.detect()
        const markers = this.localAppMarkers
        for (const marker of markers) {
            const watchPattern = new vscode.RelativePattern(this.homeDir, marker)
            const watcher = vscode.workspace.createFileSystemWatcher(watchPattern)
            watcher.onDidCreate(() => this.detect())
            watcher.onDidDelete(() => this.detect())
            this._watchers.push(watcher)
        }
    }

    public getProcessInfo(): LocalProcess {
        return {
            arch: this.arch,
            os: this.platformName,
            homeDir: this.homeDir,
            isAppInstalled: this.isInstalled,
        }
    }

    private fire(): void {
        this.onChange(this.isInstalled)
        if (this.isInstalled) {
            this.dispose()
        }
    }

    public dispose(): void {
        for (const watcher of this._watchers) {
            watcher.dispose()
        }
        this._watchers = []
    }
}
