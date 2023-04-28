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

// TODO(burmudar): use logging service to log that this has been loaded
const outputHandler = async (event: Event<string>): Promise<void> => {
    if (event.payload.startsWith('tauri:sign-in-url: ')) {
        let url = event.payload.slice('tauri:sign-in-url: '.length).trim()
        launchWithSignInUrl(url)
    }
}

// Note we currently ignore the unlisten cb returned from listen
let sidecar: String = 'sourcegraph-backend'
listen(`${sidecar}-stdout`, outputHandler)
    .then(() => console.log(`${sidecar}-stdout listener registered`))
    .catch(error => console.error(`failed to register backend-stdout handler: ${error}`))
listen(`${sidecar}-stderr`, outputHandler)
    .then(() => console.log(`${sidecar}-stderr listener registered`))
    .catch(error => console.error(`failed to register backend-stderr handler: ${error}`))
