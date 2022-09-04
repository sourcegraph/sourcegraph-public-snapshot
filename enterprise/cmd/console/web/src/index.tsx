import './index.css'

import React from 'react'
import { createRoot } from 'react-dom/client'

import { App } from './app/App'

const strictMode = true

const appElement = strictMode ? (
    <React.StrictMode>
        <App />
    </React.StrictMode>
) : (
    <App />
)

const container = document.querySelector('#root')! // eslint-disable-line @typescript-eslint/no-non-null-assertion
createRoot(container).render(appElement)
