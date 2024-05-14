import mermaid from 'mermaid'
import type { Action } from 'svelte/action'

mermaid.initialize({ startOnLoad: false })

export const renderMermaid: Action<HTMLElement, string> = (node, selector) => {
    const mermaidBlocks = node.querySelectorAll(selector)
    for (const [i, mermaidBlock] of mermaidBlocks.entries()) {
        mermaid.render(`mermaid-diagram-${i}`, mermaidBlock.textContent || '').then(({ svg }) => {
            const newDiv = document.createElement('div')
            newDiv.innerHTML = svg
            mermaidBlock.replaceWith(newDiv)
        })
    }
}
