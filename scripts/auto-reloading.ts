import chalk from 'chalk'
import io from 'socket.io'

/**
 * Returns a trigger function that notifies the extension to reload itself.
 */
export const initializeServer = () => {
    const logColor = (color: string) => (message: string) => console.log(chalk[color]('auto reload: ' + message))
    const info = logColor('blue')
    const warn = logColor('yellow')

    // Since this port is hard-coded, it must match background.tsx
    const socketIOServer = io.listen(8890)
    info('Ready for the extension to connect.')
    socketIOServer.on('connect', socket => {
        info('The extension connected.')
    })

    return () => {
        if (Object.keys(socketIOServer.clients().connected).length === 0) {
            warn('The extension has not connected yet. Try reloading it manually. Make sure that:')
            warn('- The extension is enabled')
            warn('- initializeAutoReloading() is being called in the background script')
            warn('- The extension is in developer mode (meaning you loaded it as an unpacked extension)')
        } else {
            info('Triggering a reload of the extension.')
            socketIOServer.emit('file.change', {})
        }
    }
}
