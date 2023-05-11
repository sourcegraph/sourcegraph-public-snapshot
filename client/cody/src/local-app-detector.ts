import { promises as fs } from 'fs'
import { homedir } from 'os'

const LOCAL_APP_LOCATIONS = [
    '~/Library/Application Support/sourcegraph-sp',
    '~/Library/Application Support/sourcegraph',
]

async function pathExists(path: string): Promise<boolean> {
    path = expandHomeDir(path)
    try {
        await fs.access(path)
        return true
    } catch (err) {
        return false
    }
}

function expandHomeDir(path: string): string {
    if (path.startsWith('~/')) {
        return path.replace('~', homedir())
    }
    return path
}

/**
 * Detects whether the user has the Sourcegraph app installed locally.
 */
export class LocalAppDetector {
    async detect(): Promise<boolean> {
        for (const marker of LOCAL_APP_LOCATIONS) {
            if (await pathExists(marker)) {
                return true
            }
        }
        return false
    }
}
