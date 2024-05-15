import mermaid from 'mermaid'
import type { Action } from 'svelte/action'

mermaid.initialize({ startOnLoad: false })

// renderMermaid is an action renders mermaid diagrams. It only targets descendents
// of the target element that match the provided CSS selector, and replaces them
// with a rendered SVG of that diagram.
export const renderMermaid: Action<HTMLElement, string> = (node, selector) => {
    const mermaidBlocks = node.querySelectorAll(selector)
    for (const [i, mermaidBlock] of mermaidBlocks.entries()) {
        mermaid.render(`mermaid-diagram-${i}`, mermaidBlock.textContent || '').then(({ svg }) => {
            mermaidBlock.outerHTML = svg
        })
    }
}
