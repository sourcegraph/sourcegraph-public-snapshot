// We want to polyfill first.
// prettier-ignore
import '../../app/util/polyfill'

import { injectBitbucketServer } from '../../app/bitbucket/inject'
import { injectGitHubApplication } from '../../app/github/inject'
import { injectPhabricatorApplication } from '../../app/phabricator/app'
import { injectSourcegraphApp } from '../../app/sourcegraph/inject'
import {
    isOnlySourcegraphDotCom,
    setExecuteSearchEnabled,
    setInlineSymbolSearchEnabled,
    setRenderMermaidGraphsEnabled,
    setRepositoryFileTreeEnabled,
    setServerUrls,
    setSourcegraphUrl,
    setUseCXP,
} from '../../app/util/context'
import { getURL } from '../../extension/extension'
import * as runtime from '../../extension/runtime'
import storage from '../../extension/storage'
import { StorageItems } from '../../extension/types'

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

    const handleGetStorage = (items: StorageItems) => {
        if (items.disableExtension) {
            return
        }

        // This has a default value so it should always be defined.
        // In safari, if the storage hasn't been initialized yet, this
        // will be undefined so we want to return early.
        // When safari's storage is initialized, we fire a custom event that will
        // re-run this function.
        if (!items.sourcegraphURL) {
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

        if (isOnlySourcegraphDotCom(items.serverUrls) && !items.hasSeenServerModal) {
            runtime.sendMessage({ type: 'setBadgeText', payload: '1' })
        }

        if (!isSourcegraphServer && !document.getElementById('ext-style-sheet')) {
            if (window.safari) {
                runtime.sendMessage({
                    type: 'insertCSS',
                    payload: { file: 'css/style.bundle.css', origin: window.location.origin },
                })
            } else if (isPhabricator || isGitHub || isGitHubEnterprise || isBitbucket) {
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
            setServerUrls(items.serverUrls)
            setRenderMermaidGraphsEnabled(
                items.renderMermaidGraphsEnabled === undefined ? false : items.renderMermaidGraphsEnabled
            )
            setInlineSymbolSearchEnabled(
                items.inlineSymbolSearchEnabled === undefined ? false : items.inlineSymbolSearchEnabled
            )
            setRepositoryFileTreeEnabled(
                items.repositoryFileTreeEnabled === undefined ? true : items.repositoryFileTreeEnabled
            )
            setExecuteSearchEnabled(items.executeSearchEnabled === undefined ? true : items.executeSearchEnabled)
            injectGitHubApplication(extensionMarker)
        } else if (isSourcegraphServer || /^https?:\/\/(www.)?sourcegraph.com/.test(href)) {
            setSourcegraphUrl(sourcegraphServerUrl)
            injectSourcegraphApp(extensionMarker)
        } else if (isPhabricator) {
            window.SOURCEGRAPH_PHABRICATOR_EXTENSION = true
            setSourcegraphUrl(sourcegraphServerUrl)
            injectPhabricatorApplication()
        } else if (
            document.querySelector('.bitbucket-header-logo') ||
            document.querySelector('.aui-header-logo.aui-header-logo-bitbucket')
        ) {
            setSourcegraphUrl(sourcegraphServerUrl)
            setServerUrls(items.serverUrls)
            injectBitbucketServer()
        }
        setUseCXP(items.useCXP === undefined ? false : items.useCXP)
    }

    storage.getSync(handleGetStorage)

    document.addEventListener('sourcegraph:storage-init', () => {
        storage.getSync(handleGetStorage)
    })
}

if (document.readyState === 'complete' || document.readyState === 'interactive') {
    // document is already ready to go
    injectApplication()
} else {
    document.addEventListener('DOMContentLoaded', injectApplication)
}
