import type { Mermaid } from 'mermaid'
import type { Action } from 'svelte/action'

import { uniqueID } from '$lib/dom'

// renderMermaid is an action renders mermaid diagrams. It only targets descendents
// of the target element that match the provided CSS selector, and replaces them
// with a rendered SVG of that diagram.
export const renderMermaid: Action<HTMLElement, { selector: string; isLightTheme: boolean }> = (
    node,
    { selector, isLightTheme }
) => {
    const mermaidBlocks = Array.from(node.querySelectorAll(selector), node => [node, node.textContent] as const)
    let id = uniqueID()
    let destroyed = false
    let mermaid: Mermaid | null = null

    async function render(mermaid: Mermaid, isLightTheme: boolean) {
        // It seems we have to call this every time otherwise the them won't change
        // (just calling `setConfig` didn't work)
        mermaid?.mermaidAPI.initialize({
            startOnLoad: false,
            theme: isLightTheme ? 'default' : 'dark',
        })

        for (const [i, [node, text]] of mermaidBlocks.entries()) {
            if (destroyed) break

            mermaid.mermaidAPI.render(`mermaid-diagram-${id}-${i}`, text || '').then(({ svg }) => {
                node.innerHTML = svg
            })
        }
    }

    if (mermaidBlocks.length > 0) {
        import('mermaid').then(({ default: _mermaid }) => {
            mermaid = _mermaid
            render(mermaid, isLightTheme)
        })
    }

    return {
        update(update) {
            isLightTheme = update.isLightTheme
            if (mermaid) {
                render(mermaid, update.isLightTheme)
            }
        },
        destroy() {
            destroyed = true
        },
    }
}
