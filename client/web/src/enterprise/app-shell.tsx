import { listen, Event } from '@tauri-apps/api/event'

// Sourcegraph desktop app entrypoint. There are two:
//
// * app-shell.tsx: before the Go backend has started, this is served. If the Go backend crashes,
//   then the Tauri Rust application can bring the user back here to present debugging/error handling
//   options.
// * app-main.tsx: served by the Go backend, renders the Sourcegraph web UI that you see everywhere else.

console.log('app-shell.tsx loaded')

const outputHandler = (event: Event<string>) => {
  if (event.payload.startsWith('tauri:sign-in-url: ')) {
    const url = event.payload.slice('tauri:sign-in-url: '.length).trim()
    window.location.href = url
  }
}

listen('backend-stdout', outputHandler)
listen('backend-stderr', outputHandler)
