import './index.css'

import React from 'react'
import { createRoot } from 'react-dom/client'
import { BrowserRouter } from 'react-router-dom'

import { App } from './app/App'

const appElement = (
    <React.StrictMode>
        <BrowserRouter>
            <App />
        </BrowserRouter>
    </React.StrictMode>
)
const container = document.querySelector('#root')! // eslint-disable-line @typescript-eslint/no-non-null-assertion
createRoot(container).render(appElement)
