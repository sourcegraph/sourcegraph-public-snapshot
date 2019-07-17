import '../config/polyfill'

import * as H from 'history'
import React from 'react'
import { setLinkComponent } from '../../../shared/src/components/Link'
import { injectCodeIntelligence } from '../libs/code_intelligence/inject'
import { injectExtensionMarker, NATIVE_INTEGRATION_ACTIVATED } from '../libs/sourcegraph/inject'

const IS_EXTENSION = false

setLinkComponent(({ to, children, ...props }) => (
    <a href={to && typeof to !== 'string' ? H.createPath(to) : to} {...props}>
        {children}
    </a>
))

async function init(): Promise<void> {
    console.log('Sourcegraph native integration is running')
    const sourcegraphURL = window.SOURCEGRAPH_URL
    if (!sourcegraphURL) {
        throw new Error('window.SOURCEGRAPH_URL is undefined')
    }
    const link = document.createElement('link')
    link.setAttribute('rel', 'stylesheet')
    link.setAttribute('type', 'text/css')
    link.setAttribute('href', sourcegraphURL + `/.assets/extension/css/style.bundle.css`)
    link.id = 'sourcegraph-styles'
    document.getElementsByTagName('head')[0].appendChild(link)
    window.localStorage.setItem('SOURCEGRAPH_URL', sourcegraphURL)
    window.SOURCEGRAPH_URL = sourcegraphURL
    injectExtensionMarker()
    // TODO handle subscription
    await injectCodeIntelligence(IS_EXTENSION)
    document.dispatchEvent(new CustomEvent<{}>(NATIVE_INTEGRATION_ACTIVATED))
}

init().catch(err => {
    console.error('Error initializing integration', err)
})
