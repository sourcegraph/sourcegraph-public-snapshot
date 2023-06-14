import * as vscode from 'vscode'

const LOCAL_APP_LOCATIONS: { [key: string]: string[] } = {
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
    private onChange: OnChangeCallback
    private isInstalled = false
    private localAppMarkers: string[] | undefined
    private watcher: vscode.FileSystemWatcher | undefined
    private platformName: string

    constructor(options: { onChange: OnChangeCallback }) {
        this.onChange = options.onChange
        this.platformName = process.platform
        this.localAppMarkers = LOCAL_APP_LOCATIONS[this.platformName]
        this.start()
    }

    public get isLocalAppInstalled(): boolean {
        return this.isInstalled
    }

    public async detect(): Promise<void> {
        if (!this.localAppMarkers) {
            return
        }

        if (await pathExists(this.localAppMarkers[0])) {
            if (isMac() && (await pathExists(this.localAppMarkers[1]))) {
                this.fire(true)
                return
            }
        }

        for (const marker of this.localAppMarkers) {
            const markerExists = await pathExists(expandHomeDir(marker))
            if (markerExists) {
                this.fire(true)
                return
            }
        }

        this.fire(false)
    }

    private fire(state: boolean): void {
        this.onChange(state)
        this.isInstalled = state
    }

    public start(): void {
        // Get home directory
        const homeDir = process.env.HOME
        if (this.watcher !== undefined || !homeDir) {
            return
        }
        const marker = this.localAppMarkers?.[0].replace('~/', '')
        if (!marker) {
            return
        }

        const watchPattern = new vscode.RelativePattern(homeDir, `${marker}/**`)
        this.watcher = vscode.workspace.createFileSystemWatcher(watchPattern)

        void this.detect()

        this.watcher.onDidChange(() => {
            void this.detect()
        })
        this.watcher.onDidCreate(() => {
            void this.detect()
        })
        this.watcher.onDidDelete(() => {
            void this.detect()
        })
    }

    public dispose(): void {
        if (this.watcher !== undefined) {
            this.watcher.dispose()
        }
    }
}

function isMac(): boolean {
    return process.platform === 'darwin'
}
