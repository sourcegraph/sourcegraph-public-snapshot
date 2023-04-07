import React from 'react'

import ReactDOM from 'react-dom/client'

import { PromptMixin, languagePromptMixin } from '@sourcegraph/cody-shared/src/chat/recipes/prompt-mixin'
import { WildcardThemeContext } from '@sourcegraph/wildcard'

import { App } from './App'

import './index.css'

PromptMixin.add(languagePromptMixin(navigator.language))
ReactDOM.createRoot(document.querySelector('#root') as HTMLElement).render(
    <React.StrictMode>
        <WildcardThemeContext.Provider value={{ isBranded: true }}>
            <App />
        </WildcardThemeContext.Provider>
    </React.StrictMode>
)
