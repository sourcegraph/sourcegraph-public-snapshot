import { listen, Event } from '@tauri-apps/api/event'

// Sourcegraph desktop app entrypoint. There are two:
//
// * app-shell.tsx: before the Go backend has started, this is served. If the Go backend crashes,
//   then the Tauri Rust application can bring the user back here to present debugging/error handling
//   options.
// * app-main.tsx: served by the Go backend, renders the Sourcegraph web UI that you see everywhere else.

// TODO(burmudar): use logging service to log that this has been loaded
const outputHandler = (event: Event<string>): void => {
    if (event.payload.startsWith('tauri:sign-in-url: ')) {
        const url = event.payload.slice('tauri:sign-in-url: '.length).trim()
        window.location.href = url
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
