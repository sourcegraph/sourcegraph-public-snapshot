import '../config/polyfill'

import * as H from 'history'
import React from 'react'
import { Observable } from 'rxjs'
import { startWith } from 'rxjs/operators'
import { setLinkComponent } from '../../../shared/src/components/Link'
import { determineCodeHost, injectCodeIntelligenceToCodeHost } from '../libs/code_intelligence'
import { MutationRecordLike, observeMutations } from '../shared/util/dom'

const IS_EXTENSION = false

// NOT idempotent.
async function injectModules(): Promise<void> {
    // This is added so that the browser extension doesn't
    // interfere with the native integration.
    // TODO this is racy because the script is loaded async
    const extensionMarker = document.createElement('div')
    extensionMarker.id = 'sourcegraph-app-background'
    extensionMarker.style.display = 'none'
    document.body.appendChild(extensionMarker)

    // TODO handle subscription
    const codeHost = await determineCodeHost()
    if (codeHost) {
        const mutations: Observable<MutationRecordLike[]> = observeMutations(document.body, {
            childList: true,
            subtree: true,
        }).pipe(startWith([{ addedNodes: [document.body], removedNodes: [] }]))
        await injectCodeIntelligenceToCodeHost(mutations, codeHost, IS_EXTENSION)
    }
}

setLinkComponent(({ to, children, ...props }) => (
    <a href={to && typeof to !== 'string' ? H.createPath(to) : to} {...props}>
        {children}
    </a>
))

async function init(): Promise<void> {
    const sourcegraphURL = window.SOURCEGRAPH_URL
    if (!sourcegraphURL) {
        throw new Error('window.SOURCEGRAPH_URL is undefined')
    }
    const style = document.createElement('link')
    style.setAttribute('rel', 'stylesheet')
    style.setAttribute('type', 'text/css')
    style.setAttribute('href', sourcegraphURL + `/.assets/extension/css/style.bundle.css`)
    style.id = 'sourcegraph-styles'
    document.getElementsByTagName('head')[0].appendChild(style)
    window.localStorage.setItem('SOURCEGRAPH_URL', sourcegraphURL)
    window.SOURCEGRAPH_URL = sourcegraphURL
    await injectModules()
}

init().catch(err => {
    console.error('Error initializing integration', err)
})
