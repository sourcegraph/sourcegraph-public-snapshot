import * as vscode from 'vscode'

export interface LocalProcess {
    arch?: string
    homeDir?: string | undefined
    os?: string
    isAppInstalled: boolean
}

export interface LocalAppPaths {
    [os: string]: {
        dir: string
        file: string
    }[]
}

const LOCAL_APP_LOCATIONS: LocalAppPaths = {
    // Only Apple silicon is supported
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
    private localAppMarkers
    private appFsPaths: string[] = []

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
        this.init()
    }

    private init(): void {
        // if conditions are not met, this will be a noop
        if (this._watchers.length || !this.isSupported || !this.homeDir || this.isInstalled) {
            return
        }
        // Create filePaths and file watchers
        const makers = this.localAppMarkers
        for (const maker of makers) {
            const dirPath = this.expandHomeDir(maker.dir)
            const dirUri = vscode.Uri.file(dirPath)
            const watchPattern = new vscode.RelativePattern(dirUri, maker.file)
            const watcher = vscode.workspace.createFileSystemWatcher(watchPattern)
            watcher.onDidChange(() => this.detect())
            this._watchers.push(watcher)
            this.appFsPaths.push(dirPath + maker.file)
        }
        void this.detect()
    }

    private async detect(): Promise<void> {
        if (!this.isSupported || !this.appFsPaths.length) {
            return
        }
        if (await Promise.any(this.appFsPaths.map(file => pathExists(file)))) {
            this.fire()
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

    private expandHomeDir(path: string): string {
        if (path.startsWith('~/')) {
            return path.replace('~', process.env.HOME || '')
        }
        return path
    }

    // We can dispose the file watcher when app is found or when user has logged in
    private fire(): void {
        console.info('app found')
        this.isInstalled = true
        this.onChange(this.isInstalled)
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
