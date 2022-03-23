import vscode from 'vscode'

import { INSTANCE_VERSION_NUMBER_KEY, LocalStorageService } from './LocalStorageService'

export async function displayWarning(warning: string): Promise<void> {
    await vscode.window.showErrorMessage(warning)
}

export function displayInstanceVersionWarnings(localStorageService: LocalStorageService): void {
    const versionNumber = localStorageService.getValue(INSTANCE_VERSION_NUMBER_KEY)
    if (!versionNumber) {
        displayWarning('Cannot determine instance version number').catch(() => {})
    }
    if (versionNumber < '3320') {
        displayWarning(
            'Your Sourcegraph instance version is not fully compatible with the Sourcegraph extension. Please ask your site admin to upgrade to version 3.32.0 or above. Read more about version support in our [troubleshooting docs](https://docs.sourcegraph.com/admin/how-to/troubleshoot-sg-extension#unsupported-features-by-sourcegraph-version).'
        ).catch(() => {})
    }
    return
}
