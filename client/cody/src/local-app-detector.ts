import { promises as fs } from 'fs'
import { homedir } from 'os'

import { Disposable } from 'vscode'

const INTERVAL = 5000

const LOCAL_APP_LOCATIONS = ['~/Library/Application Support/sourcegraph', '~/Applications/Sourcegraph App.app']

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

    constructor(options: { onChange: OnChangeCallback }) {
        this.onChange = options.onChange
    }

    public async detect(): Promise<void> {
        let detected = false
        for (const marker of LOCAL_APP_LOCATIONS) {
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
