import { listen, Event } from '@tauri-apps/api/event'
import { invoke } from '@tauri-apps/api/tauri'

function addRedirectParamToSignInUrl(url: string, returnTo: string) {
    const urlObject = new URL(url)
    urlObject.searchParams.append('redirect', returnTo)
    return urlObject.toString()
}

async function getLaunchPathFromTauri(): Promise<string> {
    return (await invoke('get_launch_path')) as string
}

async function launchWithSignInUrl(url: string) {
    const launchPath = await getLaunchPathFromTauri()
    if (launchPath) {
        console.log('Using launch path:', launchPath)
        url = addRedirectParamToSignInUrl(url, launchPath)
    }
    window.location.href = url
}

// Sourcegraph desktop app entrypoint. There are two:
//
// * app-shell.tsx: before the Go backend has started, this is served. If the Go backend crashes,
//   then the Tauri Rust application can bring the user back here to present debugging/error handling
//   options.
// * app-main.tsx: served by the Go backend, renders the Sourcegraph web UI that you see everywhere else.

interface TauriLog {
    level: number
    message: string
}

// TODO(burmudar): use logging service to log that this has been loaded
const outputHandler = (event: Event<TauriLog>): void => {
    if (event.payload.message.includes('tauri:sign-in-url: ')) {
        const url = event.payload.message.split('tauri:sign-in-url: ')[1]
        launchWithSignInUrl(url)
    }
}

// Note we currently ignore the unlisten cb returned from listen
listen('log://log', outputHandler)
    .then(() => console.log('registered stdout handler'))
    .catch(error => console.error(`failed to register stdout handler: ${error}`))
