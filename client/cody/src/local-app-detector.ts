import { promises as fs } from 'fs'
import { homedir, platform } from 'os'

import { Disposable } from 'vscode'

const INTERVAL = 10000

const LOCAL_APP_LOCATIONS: { [key: string]: string[] } = { darwin: ['~/Library/Application Support/sourcegraph'] }

async function pathExists(path: string): Promise<boolean> {
    path = expandHomeDir(path)
    try {
        await fs.access(path)
        return true
    } catch {
        return false
    }
}

function expandHomeDir(path: string): string {
    if (path.startsWith('~/')) {
        return path.replace('~', homedir())
    }
    return path
}

type OnChangeCallback = (value: boolean) => void

/**
 * Detects whether the user has the Sourcegraph app installed locally.
 */
export class LocalAppDetector implements Disposable {
    private intervalHandle: ReturnType<typeof setInterval> | undefined
    private onChange: OnChangeCallback
    private lastState = false
    private localAppMarkers: string[] | undefined

    constructor(options: { onChange: OnChangeCallback }) {
        this.onChange = options.onChange
        const platformName = platform()
        this.localAppMarkers = LOCAL_APP_LOCATIONS[platformName]
    }

    public async detect(): Promise<void> {
        let detected = false
        if (!this.localAppMarkers) {
            return
        }

        for (const marker of this.localAppMarkers) {
            if (await pathExists(marker)) {
                detected = true
                break
            }
        }

        if (detected !== this.lastState) {
            this.lastState = detected
            this.onChange(detected)
        }
    }

    public start(): void {
        if (this.intervalHandle !== undefined) {
            return
        }

        this.intervalHandle = setInterval(() => {
            void this.detect()
        }, INTERVAL)

        void this.detect()
    }

    public stop(): void {
        if (this.intervalHandle !== undefined) {
            clearInterval(this.intervalHandle)
        }
    }

    public dispose(): void {
        this.stop()
    }
}
