import { callJava, setDarkMode } from './call-java-mock'
import { renderColorDebugger } from './renderColorDebugger'

const iframeNode = document.querySelector('#webview') as HTMLIFrameElement

// Initialize app for standalone server
iframeNode.addEventListener('load', () => {
    const iframeWindow = iframeNode.contentWindow
    if (iframeWindow !== null) {
        iframeWindow.callJava = callJava
        iframeWindow
            .initializeSourcegraph()
            .then(() => {})
            .catch(() => {})
    }
})

// Detect dark or light mode preference
if (window.matchMedia('(prefers-color-scheme: dark)').matches) {
    setDarkMode(true)
    // eslint-disable-next-line @typescript-eslint/no-non-null-assertion
    document.body.parentElement!.className = 'dark'
}

// Render the theme color debugger when the URL contains `?color-debug`
if (location.href.includes('color-debug')) {
    renderColorDebugger()
}
