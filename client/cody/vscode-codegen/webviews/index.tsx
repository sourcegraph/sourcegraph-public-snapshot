import React from 'react'

import ReactDOM from 'react-dom/client'

import App from './App'

import './index.css'

ReactDOM.createRoot(document.querySelector('#root') as HTMLElement).render(
    <React.StrictMode>
        <App />
    </React.StrictMode>
)
