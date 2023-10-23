import '@sourcegraph/shared/src/polyfills'

import { createRoot } from 'react-dom/client'

import { logger } from '@sourcegraph/common'

import { initAppShell } from '../../storm/app-shell-init'

import { EmbeddedWebApp } from './EmbeddedWebApp'

const appShellPromise = initAppShell()

// It's important to have a root component in a separate file to create a react-refresh boundary and avoid page reload.
window.addEventListener('DOMContentLoaded', async () => {
    const root = createRoot(document.querySelector('#root')!)

    try {
        const { graphqlClient } = await appShellPromise

        root.render(<EmbeddedWebApp graphqlClient={graphqlClient} />)
    } catch (error) {
        logger.error('Failed to initialize the app shell', error)
    }
})
