import mermaid from 'mermaid'
import type { Action } from 'svelte/action'

// renderMermaid is an action renders mermaid diagrams. It only targets descendents
// of the target element that match the provided CSS selector, and replaces them
// with a rendered SVG of that diagram.
export const renderMermaid: Action<HTMLElement, { selector: string; isLightTheme: boolean }> = (
    node,
    { selector, isLightTheme }
) => {
    // TODO(camdencheek): this does not update when the theme changes. We should
    // manually specify the theme and have it react to the global CSS props.
    mermaid.mermaidAPI.initialize({
        startOnLoad: false,
        theme: isLightTheme ? 'default' : 'dark',
    })
    const mermaidBlocks = node.querySelectorAll(selector)
    for (const [i, mermaidBlock] of mermaidBlocks.entries()) {
        mermaid.mermaidAPI.render(`mermaid-diagram-${i}`, mermaidBlock.textContent || '').then(({ svg }) => {
            mermaidBlock.outerHTML = svg
        })
    }
}
