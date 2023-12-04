import { listen, type Event } from '@tauri-apps/api/event'
import { invoke } from '@tauri-apps/api/tauri'

import { logger } from '@sourcegraph/common'

// Sourcegraph desktop app entrypoint. There are two:
//
// * app-shell.tsx: before the Go backend has started, this is served. If the Go backend crashes,
//   then the Tauri Rust application can bring the user back here to present debugging/error handling
//   options.
// * app/main.tsx: served by the Go backend, renders the Sourcegraph web UI that you see everywhere else.

function addRedirectParamToSignInUrl(url: string, returnTo: string): string {
    const urlObject = new URL(url)
    urlObject.searchParams.append('redirect', returnTo)
    return urlObject.toString()
}

async function getLaunchPathFromTauri(): Promise<string> {
    return invoke('get_launch_path')
}

async function launchWithSignInUrl(url: string): Promise<void> {
    const launchPath = await getLaunchPathFromTauri()
    if (launchPath) {
        logger.log('Using launch path:', launchPath)
        url = addRedirectParamToSignInUrl(url, launchPath)
    }
    window.location.href = url
}

interface AppShellReadyPayload {
    sign_in_url: string
}

const appShellReady = (payload: AppShellReadyPayload): void => {
    if (!payload) {
        return
    }
    logger.log('app-shell-ready', payload)
    launchWithSignInUrl(payload.sign_in_url).catch(error =>
        // eslint-disable-next-line no-console
        console.error(`failed to launch with sign-in URL: ${error}`)
    )
}

listen('app-shell-ready', (event: Event<AppShellReadyPayload>) => appShellReady(event.payload))
    .then(() => logger.log('registered app-shell-ready handler'))
    .catch(error => logger.error(`failed to register app-shell-ready handler: ${error}`))

await invoke('app_shell_loaded')
    .then(payload => appShellReady(payload as AppShellReadyPayload))
    .catch(error => logger.error(`failed to inform Tauri app_shell_loaded: ${error}`))
