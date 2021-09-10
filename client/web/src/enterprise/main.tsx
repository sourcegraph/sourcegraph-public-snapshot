/// <reference types="react/next" />
/// <reference types="react-dom/next" />

// This is the entry point for the enterprise web app

// Order is important here
// Don't remove the empty lines between these imports

import '@sourcegraph/shared/src/polyfills'

import '../sentry'

import { ApolloProvider } from '@apollo/client'
import { createBrowserHistory } from 'history'
import React from 'react'
import ReactDOM from 'react-dom'
import { BrowserRouter } from 'react-router-dom'

import { getWebGraphQLClient } from '../backend/graphql'

import { EnterpriseWebApp } from './EnterpriseWebApp'

// It's important to have a root component in a separate file to create a react-refresh boundary and avoid page reload.
// https://github.com/pmmmwh/react-refresh-webpack-plugin/blob/main/docs/TROUBLESHOOTING.md#edits-always-lead-to-full-reload
// window.addEventListener('DOMContentLoaded', () => {
const root = document.querySelector('#root')!
// TODO(sqs): <React.StrictMode> causes many problems currently, fix those! then wrap the app in <React.StrictMode>.

const graphQLClient = getWebGraphQLClient()

const jsx = (
    <ApolloProvider client={graphQLClient}>
        <BrowserRouter>
            <EnterpriseWebApp history={createBrowserHistory()} graphQLClient={graphQLClient} />
        </BrowserRouter>
    </ApolloProvider>
)

if (!location.search.includes('noscript') && localStorage.getItem('noscript') === null) {
    const hydrate = root.hasChildNodes()
    if (hydrate && !location.search.includes('nohydrate')) {
        window.__waitForAuthUser.then(() => ReactDOM.hydrateRoot(root, jsx, { unstable_strictMode: true }))
    } else {
        ReactDOM.createRoot(root).render(jsx)
    }
}
// })
