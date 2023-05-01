import { listen, Event } from '@tauri-apps/api/event'

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
        window.location.href = url
    }
}

// Note we currently ignore the unlisten cb returned from listen
listen('log://log', outputHandler)
    .then(() => console.log('registered stdout handler'))
    .catch(error => console.error(`failed to register stdout handler: ${error}`))
