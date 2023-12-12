// Sourcegraph desktop app entrypoint. There are two:
//
// * app-shell.tsx: before the Go backend has started, this is served. If the Go backend crashes,
//   then the Tauri Rust application can bring the user back here to present debugging/error handling
//   options.
// * app-main.tsx: served by the Go backend, renders the Sourcegraph web UI that you see everywhere else.

// Order is important here
// Don't remove the empty lines between these imports

// prettier-ignore-start
import '@sourcegraph/shared/src/polyfills'
// prettier-ignore-end

import '../initBuildInfo'
import '../monitoring/initMonitoring'

import { createRoot } from 'react-dom/client'

import { logger } from '@sourcegraph/common'
import { setLinkComponent } from '@sourcegraph/wildcard'

import { TauriLink } from '../app/TauriLink'
import { initAppShell } from '../storm/app-shell-init'

import { EnterpriseWebApp } from './EnterpriseWebApp'

const appShellPromise = initAppShell()

setLinkComponent(TauriLink)

// It's important to have a root component in a separate file to create a react-refresh boundary and avoid page reload.
// https://github.com/pmmmwh/react-refresh-webpack-plugin/blob/main/docs/TROUBLESHOOTING.md#edits-always-lead-to-full-reload
window.addEventListener('DOMContentLoaded', async () => {
    const root = createRoot(document.querySelector('#root')!)

    try {
        const { graphqlClient, temporarySettingsStorage, telemetryRecorder } = await appShellPromise

        root.render(
            <EnterpriseWebApp
                graphqlClient={graphqlClient}
                temporarySettingsStorage={temporarySettingsStorage}
                telemetryRecorder={telemetryRecorder}
            />
        )
    } catch (error) {
        logger.error('Failed to initialize the app shell', error)
    }
})

if (process.env.DEV_WEB_BUILDER === 'esbuild' && process.env.NODE_ENV === 'development') {
    new EventSource('/.assets/esbuild').addEventListener('change', () => {
        location.reload()
    })
}
