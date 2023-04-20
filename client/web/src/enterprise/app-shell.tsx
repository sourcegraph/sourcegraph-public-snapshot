console.log('app-shell.tsx loaded')

import { listen } from '@tauri-apps/api/event'

const outputHandler = (event) => {
    console.log(':: ' + event.payload)
    if (event.payload.startsWith('tauri:sign-in-url: ')) {
        const url = event.payload.slice('tauri:sign-in-url: '.length).trim()
        window.location.href = url
    }
}

listen('backend-stdout', outputHandler)
listen('backend-stderr', outputHandler)
