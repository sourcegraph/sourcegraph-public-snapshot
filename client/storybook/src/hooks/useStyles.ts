import { useEffect, useMemo } from 'react'

const createStyleTag = (id: string): HTMLStyleElement => {
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
 * @returns The created `<style>` tag
 */
export const useStyles = (css: string): HTMLStyleElement => {
    const styleTag = useMemo(() => {
        const styleTag = document.querySelector<HTMLStyleElement>('story-styles') || createStyleTag('story-styles')
        styleTag.textContent = css
        return styleTag
    }, [css])

    useEffect(() => () => styleTag.remove(), [styleTag])

    return styleTag
}
