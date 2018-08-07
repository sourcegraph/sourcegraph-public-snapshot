import { injectPhabricatorBlobAnnotators } from './inject'
import { expanderListen, javelinPierce, metaClickOverride, setupPageLoadListener } from './util'

// This is injection for the chrome extension.
export function injectPhabricatorApplication(): void {
    // make sure this is called before javelinPierce
    const inject = () => {
        javelinPierce(expanderListen, 'body')
        javelinPierce(metaClickOverride, 'body')

        injectModules()
    }

    if (document.readyState === 'complete') {
        injectModules()
    } else {
        document.addEventListener('phabPageLoaded', inject)
        javelinPierce(setupPageLoadListener, 'body')
    }
}

function injectModules(): void {
    injectPhabricatorBlobAnnotators().catch(e => console.error(e))
}
