/// <reference types="react/next" />
/// <reference types="react-dom/next" />

import './browserEnv'

import { ApolloProvider, NormalizedCacheObject } from '@apollo/client'
import { getDataFromTree, renderToStringWithData } from '@apollo/client/react/ssr'
import { createMemoryHistory } from 'history'
import React from 'react'
import ReactDOMServer, { renderToString } from 'react-dom/server'
import { StaticRouter } from 'react-router'

// TODO(sqs): separate into enterprise/oss
import { getSSRGraphQLClient } from '@sourcegraph/web/src/backend/graphql'
import { EnterpriseWebApp } from '@sourcegraph/web/src/enterprise/EnterpriseWebApp'

export interface RenderRequest {
    requestURI: string
    jscontext: object
}

export interface RenderResponse {
    html?: string
    initialState?: NormalizedCacheObject
    redirectURL?: string
    error?: string
}

export const render = async ({ requestURI, jscontext }: RenderRequest): Promise<RenderResponse> => {
    console.log()
    console.log('####################################################################################')
    console.log(`# ${requestURI}`)

    // TODO(sqs): not parallel-safe
    if (jscontext && Object.keys(jscontext) > 0 /* TODO(sqs): remove this check, just for curl debugging */) {
        global.window.context = jscontext
    }
    global.window.context.PRERENDER = true

    // Pre-fetch queries.
    const graphQLClient = getSSRGraphQLClient()

    const routerContext: { url?: string } = {}
    const history = createMemoryHistory({})
    const url = new URL(requestURI, 'https://example.com')
    history.location = { pathname: url.pathname, search: url.search, hash: url.hash, state: undefined }
    const app = (
        // TODO(sqs): wrap in <React.StrictMode>
        <ApolloProvider client={graphQLClient}>
            <StaticRouter location={requestURI} context={routerContext}>
                <EnterpriseWebApp history={history} graphQLClient={graphQLClient} />
            </StaticRouter>
        </ApolloProvider>
    )

    const html = await renderToStringWithData(app)
    const initialState = graphQLClient.extract()

    const html2 = renderToString(app)
    console.log('1111111111', html)
    console.log('2222222222', html2)

    const result = {
        html: html2,
        redirectURL: routerContext.url,
        initialState,
    }

    console.log(`# HTML: ${JSON.stringify(html)}`)
    if (result.redirectURL) {
        console.log(`# REDIRECT ${result.redirectURL}`)
    }
    console.log('####################################################################################')
    console.log()

    return result
}

if (false) {
    render({ requestURI: '/', jscontext: {} })
        .then(response => console.log('ZZ', response))
        .catch(error => console.error('Error:', error))
        .finally(() => {
            console.log('EXIT111')
            process.exit(0)
        })
}
