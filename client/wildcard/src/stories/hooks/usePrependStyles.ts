import { useEffect, useMemo } from 'react'

const prependStyleTag = (id: string): HTMLStyleElement => {
    const styleTag = document.createElement('style')
    styleTag.id = id

    // Prepend global CSS styles to document head to keep them before CSS modules
    document.head.prepend(styleTag)

    return styleTag
}

/**
 * Apply a CSS string to the document head
 *
 * @param css Stringified CSS to inject into a `<style>` tag
 */
export const usePrependStyles = (styleTagId: string, css?: string): void => {
    const styleTag = useMemo(() => {
        if (!css) {
            return undefined
        }

        const styleTag = document.querySelector<HTMLStyleElement>(styleTagId) || prependStyleTag(styleTagId)
        styleTag.textContent = css

        return styleTag
    }, [styleTagId, css])

    useEffect(() => () => styleTag?.remove(), [styleTag])
}
