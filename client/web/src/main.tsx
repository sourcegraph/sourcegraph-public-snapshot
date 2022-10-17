// This is the entry point for the web app

// Order is important here
// Don't remove the empty lines between these imports

import '@sourcegraph/shared/src/polyfills'

import './initBuildInfo'
import './monitoring/initMonitoring'

import { createRoot } from 'react-dom/client'

import { OpenSourceWebApp } from './OpenSourceWebApp'

// It's important to have a root component in a separate file to create a react-refresh boundary and avoid page reload.
// https://github.com/pmmmwh/react-refresh-webpack-plugin/blob/main/docs/TROUBLESHOOTING.md#edits-always-lead-to-full-reload
window.addEventListener('DOMContentLoaded', () => {
    const root = createRoot(document.querySelector('#root')!)

    root.render(<OpenSourceWebApp />)
})
