import React from 'react'

import ReactDOM from 'react-dom/client'

import { WildcardThemeContext } from '@sourcegraph/wildcard'

import { App } from './App'

import './index.css'

ReactDOM.createRoot(document.querySelector('#root') as HTMLElement).render(
    <React.StrictMode>
        <WildcardThemeContext.Provider value={{ isBranded: true }}>
            <App />
        </WildcardThemeContext.Provider>
    </React.StrictMode>
)
