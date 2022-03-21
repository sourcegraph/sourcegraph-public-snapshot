import vscode from 'vscode'

import { INSTANCE_VERSION_NUMBER_KEY, LocalStorageService } from './LocalStorageService'

export async function displayWarning(warning: string): Promise<void> {
    await vscode.window.showErrorMessage(warning)
}

export function instanceVersionWarnings(localStorageService: LocalStorageService): void {
    const versionNumber = localStorageService.getValue(INSTANCE_VERSION_NUMBER_KEY)
    if (!versionNumber) {
        displayWarning('Cannot determine instance version number').catch(() => {})
    }
    if (versionNumber < '3320') {
        displayWarning(
            'Your Sourcegraph instance version is not fully compatible with the Sourcegraph extension. Please ask your site admin to upgrade the Sourcegraph instance to 3.32.0 or above for full compatibility. Visit our [troubleshooting docs](https://docs.sourcegraph.com/admin/how-to/troubleshoot-sg-extension#unsupported-features-by-sourcegraph-version) to learn more about support for your Sourcegraph instance.'
        ).catch(() => {})
    }
    return
}
