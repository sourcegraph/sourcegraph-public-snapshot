import io from 'socket.io-client'

/**
 * Reloads the extension when notified from the development server. Only enabled
 * during development when `process.env.AUTO_RELOAD !== 'false'.
 */
const initializeExtension = () =>
    chrome.management.getSelf(self => {
        if (self.installType === 'development') {
            // Since the port is hard-coded, it must match scripts/dev.ts
            io.connect('http://localhost:8890').on('file.change', () => chrome.runtime.reload())
        }
    })

initializeExtension()
