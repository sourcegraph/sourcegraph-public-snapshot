import mermaid from 'mermaid'
import * as React from 'react'
import { render } from 'react-dom'
import storage from '../../browser/storage'
import { Alerts } from '../../shared/components/Alerts'
import { ConfigureSourcegraphButton } from '../../shared/components/ConfigureSourcegraphButton'
import { ContextualSourcegraphButton } from '../../shared/components/ContextualSourcegraphButton'
import { ServerAuthButton } from '../../shared/components/ServerAuthButton'
import { SymbolsDropdownContainer } from '../../shared/components/SymbolsDropdownContainer'
import { WithResolvedRev } from '../../shared/components/WithResolvedRev'
import { hideTooltip } from '../../shared/repo/tooltips'
import { inlineSymbolSearchEnabled, renderMermaidGraphsEnabled } from '../../shared/util/context'
import { getFileContainers, parseURL } from './util'

async function refreshModules(): Promise<void> {
    for (const el of Array.from(document.getElementsByClassName('sourcegraph-app-annotator'))) {
        el.remove()
    }
    for (const el of Array.from(document.getElementsByClassName('sourcegraph-app-annotator-base'))) {
        el.remove()
    }
    for (const el of Array.from(document.querySelectorAll('.sg-annotated'))) {
        el.classList.remove('sg-annotated')
    }
    hideTooltip()
    await inject()
}

window.addEventListener('pjax:end', async () => {
    await refreshModules()
})

export async function injectGitHubApplication(marker: HTMLElement): Promise<void> {
    document.body.appendChild(marker)
    await inject()
}

async function inject(): Promise<void> {
    injectServerBanner()
    injectOpenOnSourcegraphButton()

    injectMermaid()

    injectInlineSearch()
}

function injectServerBanner(): void {
    if (window.localStorage['server-banner-enabled'] !== 'true') {
        return
    }

    const { isPullRequest, repoName } = parseURL()
    if (!isPullRequest) {
        return
    }
    // Check which files were modified.
    const files = getFileContainers()
    if (!files.length) {
        return
    }

    let mount = document.getElementById('server-alert-mount')
    if (!mount) {
        mount = document.createElement('div')
        mount.id = 'server-alert-mount'
        const container = document.getElementById('partial-discussion-header')
        if (!container) {
            return
        }
        container.appendChild(mount)
    }
    render(<Alerts repoName={repoName} />, mount)
}

/**
 * Appends an Open on Sourcegraph button to the GitHub DOM.
 * The button is only rendered on a repo homepage after the "find file" button.
 */
function injectOpenOnSourcegraphButton(): void {
    storage.getSync(items => {
        const container = createOpenOnSourcegraphIfNotExists()

        if (items.featureFlags.useExtensions) {
            container.classList.add('use-extensions')
        }

        const pageheadActions = document.querySelector('.pagehead-actions')
        if (!pageheadActions || !pageheadActions.children.length) {
            return
        }
        pageheadActions.insertBefore(container, pageheadActions.children[0])
        if (container) {
            const { repoName, rev } = parseURL()
            if (repoName) {
                render(
                    <WithResolvedRev
                        component={ContextualSourcegraphButton}
                        repoName={repoName}
                        rev={rev}
                        defaultBranch={'HEAD'}
                        notFoundComponent={ConfigureSourcegraphButton}
                        requireAuthComponent={ServerAuthButton}
                    />,
                    container
                )
            }
        }
    })
}

function injectMermaid(): void {
    if (!renderMermaidGraphsEnabled) {
        return
    }

    // The structure looks like:
    //
    //    ...
    //    <pre lang="mermaid">
    //       <code>
    //          graph TD;
    //             A-->B;
    //       </code>
    //    </pre>
    //   ...
    //
    // We want to end up with:
    //
    //    ...
    //    <pre lang="mermaid">
    //       <code>
    //          graph TD;
    //             A-->B;
    //       </code>
    //    </pre>
    //    <svg>
    //       /* SVG FROM MERMAID GOES HERE */
    //    </svg>
    //   ...

    let id = 1

    const renderMermaidCharts = () => {
        const pres = document.querySelectorAll('pre[lang=mermaid]')
        for (const pre of pres) {
            const el = pre as HTMLElement
            if (el.style.display === 'none') {
                // already rendered
                continue
            }
            el.style.display = 'none'
            const chartDefinition = pre.getElementsByTagName('code')[0].textContent || ''
            const chartID = `mermaid_${id++}`
            mermaid.mermaidAPI.render(chartID, chartDefinition, svg => el.insertAdjacentHTML('afterend', svg))
        }
    }

    // Render mermaid charts async and debounce the rendering
    // to minimize impact on page load.
    let timeout: number | undefined
    const handleDomChange = () => {
        clearTimeout(timeout)
        // Need to use window.setTimeout because:
        // https://github.com/DefinitelyTyped/DefinitelyTyped/issues/21310#issuecomment-367919251
        timeout = window.setTimeout(() => renderMermaidCharts(), 200)
    }

    const observer = new MutationObserver(() => handleDomChange())
    observer.observe(document.body, { subtree: true, childList: true })
    handleDomChange()
}

function injectInlineSearch(): void {
    if (!inlineSymbolSearchEnabled) {
        return
    }

    // idempotently create a div to render the autocomplete react component inside of
    function createAutoCompleteContainerMount(textArea: HTMLTextAreaElement): HTMLDivElement | undefined {
        const parentDiv = textArea.parentElement
        if (!parentDiv) {
            return undefined
        }

        const className = 'symbols-autocomplete'

        const existingMount = parentDiv.querySelector(`.${className}`) as HTMLDivElement | null
        if (existingMount) {
            return existingMount
        }

        const mountElement = document.createElement('div')
        mountElement.className = className
        parentDiv.appendChild(mountElement)

        return mountElement
    }

    // lazily attach the symbols dropdown container whenever
    // a text area is focused
    document.addEventListener('focusin', e => {
        if (!e.target) {
            return
        }

        const target = e.target as HTMLElement

        if (target.tagName !== 'TEXTAREA') {
            return
        }

        const textArea = target as HTMLTextAreaElement
        const mountElement = createAutoCompleteContainerMount(textArea)
        if (mountElement) {
            render(<SymbolsDropdownContainer textBoxRef={textArea} />, mountElement)
        }
    })
}

const OPEN_ON_SOURCEGRAPH_ID = 'open-on-sourcegraph'

function createOpenOnSourcegraphIfNotExists(): HTMLElement {
    let container = document.getElementById(OPEN_ON_SOURCEGRAPH_ID)
    if (container) {
        container.remove()
    }

    container = document.createElement('li')
    container.id = OPEN_ON_SOURCEGRAPH_ID
    return container
}
