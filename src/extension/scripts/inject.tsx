// We want to polyfill first.
// prettier-ignore
import '../../config/polyfill'

import { getURL } from '../../browser/extension'
import * as runtime from '../../browser/runtime'
import storage from '../../browser/storage'
import { StorageItems } from '../../browser/types'
import {
    setInlineSymbolSearchEnabled,
    setRenderMermaidGraphsEnabled,
    setSourcegraphUrl,
    setUseExtensions,
} from '../../shared/util/context'
import { featureFlags } from '../../shared/util/featureFlags'

import { injectBitbucketServer } from '../../libs/bitbucket/inject'
import { injectCodeIntelligence } from '../../libs/code_intelligence'
import { injectGitHubApplication } from '../../libs/github/inject'
import { checkIsGitlab } from '../../libs/gitlab/code_intelligence'
import { injectSourcegraphApp } from '../../libs/sourcegraph/inject'
import { assertEnv } from '../envAssertion'

assertEnv('CONTENT')

/**
 * Main entry point into browser extension.
 */
function injectApplication(): void {
    const extensionMarker = document.createElement('div')
    extensionMarker.id = 'sourcegraph-app-background'
    extensionMarker.style.display = 'none'
    if (document.getElementById(extensionMarker.id)) {
        return
    }

    const href = window.location.href

    const handleGetStorage = async (items: StorageItems) => {
        if (items.disableExtension) {
            return
        }

        const srcgEl = document.getElementById('sourcegraph-chrome-webstore-item')
        const sourcegraphServerUrl = items.sourcegraphURL || 'https://sourcegraph.com'
        const isSourcegraphServer = window.location.origin === sourcegraphServerUrl || !!srcgEl
        const isPhabricator =
            Boolean(document.querySelector('.phabricator-wordmark')) &&
            Boolean(items.enterpriseUrls.find(url => url === window.location.origin))

        const isGitHub = /^https?:\/\/(www.)?github.com/.test(href)
        const ogSiteName = document.head.querySelector(`meta[property='og:site_name']`) as HTMLMetaElement
        const isGitHubEnterprise = ogSiteName ? ogSiteName.content === 'GitHub Enterprise' : false
        const isBitbucket =
            document.querySelector('.bitbucket-header-logo') ||
            document.querySelector('.aui-header-logo.aui-header-logo-bitbucket')
        const isGitlab = checkIsGitlab()

        if (!isSourcegraphServer && !document.getElementById('ext-style-sheet')) {
            if (window.safari) {
                runtime.sendMessage({
                    type: 'insertCSS',
                    payload: { file: 'css/style.bundle.css', origin: window.location.origin },
                })
            } else if (isPhabricator || isGitHub || isGitHubEnterprise || isBitbucket || isGitlab) {
                const styleSheet = document.createElement('link') as HTMLLinkElement
                styleSheet.id = 'ext-style-sheet'
                styleSheet.rel = 'stylesheet'
                styleSheet.type = 'text/css'
                styleSheet.href = getURL('css/style.bundle.css')
                document.head.appendChild(styleSheet)
            }
        }

        if (isGitHub || isGitHubEnterprise) {
            setSourcegraphUrl(sourcegraphServerUrl)
            setRenderMermaidGraphsEnabled(
                items.renderMermaidGraphsEnabled === undefined ? false : items.renderMermaidGraphsEnabled
            )
            setInlineSymbolSearchEnabled(
                items.inlineSymbolSearchEnabled === undefined ? false : items.inlineSymbolSearchEnabled
            )
            injectGitHubApplication(extensionMarker)
        } else if (isSourcegraphServer || /^https?:\/\/(www.)?sourcegraph.com/.test(href)) {
            setSourcegraphUrl(sourcegraphServerUrl)
            injectSourcegraphApp(extensionMarker)
        } else if (isPhabricator) {
            window.SOURCEGRAPH_PHABRICATOR_EXTENSION = true
            setSourcegraphUrl(sourcegraphServerUrl)
        } else if (
            document.querySelector('.bitbucket-header-logo') ||
            document.querySelector('.aui-header-logo.aui-header-logo-bitbucket')
        ) {
            setSourcegraphUrl(sourcegraphServerUrl)
            injectBitbucketServer()
        }

        if (isGitHub || isPhabricator || isGitlab) {
            if (isGitlab || (await featureFlags.isEnabled('newInject'))) {
                const subscriptions = await injectCodeIntelligence()
                window.addEventListener('unload', () => subscriptions.unsubscribe())
            }
        }

        setUseExtensions(items.useExtensions === undefined ? false : items.useExtensions)
    }

    storage.getSync(handleGetStorage)

    document.addEventListener('sourcegraph:storage-init', () => {
        storage.getSync(handleGetStorage)
    })
    // Allow users to set this via the console.
    ;(window as any).sourcegraphFeatureFlags = featureFlags
}

if (document.readyState === 'complete' || document.readyState === 'interactive') {
    // document is already ready to go
    injectApplication()
} else {
    document.addEventListener('DOMContentLoaded', injectApplication)
}
