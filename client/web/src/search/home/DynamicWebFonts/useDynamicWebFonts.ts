import { useEffect, useState } from 'react'

import { logger } from '@sourcegraph/common'

/**
 * Add `document.fonts` to Typescript type definitions based on the issue:
 * https://github.com/Microsoft/TypeScript/issues/30984#issuecomment-631991019
 */
declare global {
    type CSSOMString = string

    interface FontFaceDescriptors {
        family?: CSSOMString
        style?: CSSOMString
        weight?: CSSOMString
        stretch?: CSSOMString
        unicodeRange?: CSSOMString
        variant?: CSSOMString
        featureSettings?: CSSOMString
    }
}

export interface DynamicWebFont extends FontFaceDescriptors {
    family: string
    source: string
}

export function useDynamicWebFonts(fonts: DynamicWebFont[]): boolean {
    // The font size parameter value is not important for our use case and it can be set to any value
    const areFontsLoadedInitially = fonts.every(font => document.fonts.check(`12px ${font.family}`))
    const [areFontsLoaded, setAreFontsLoaded] = useState(areFontsLoadedInitially)
    const [hasNetworkError, setHasNetworkError] = useState(false)

    useEffect(() => {
        async function loadFonts(): Promise<void> {
            const loadingFonts = fonts.map(font => {
                const { family, source, ...fontDescriptors } = font

                // Use `FontFace()` API to create font instances.
                // https://developer.mozilla.org/en-US/docs/Web/API/FontFace/FontFace
                const fontFace = new FontFace(family, source, { ...fontDescriptors, family })

                // Load font using CSS Font Loading Module Level 3:
                // https://drafts.csswg.org/css-font-loading/#font-face-constructor
                return fontFace.load()
            })

            await Promise.all(loadingFonts).then(loadedFonts => {
                for (const font of loadedFonts) {
                    document.fonts.add(font)
                }

                setHasNetworkError(false)
                setAreFontsLoaded(true)
            })
        }

        // If fonts are available, skip redundant network request and proceed to UI rendering.
        if (!areFontsLoaded) {
            loadFonts().catch(error => {
                logger.error(error)
                setHasNetworkError(true)
            })
        }
    }, [fonts, areFontsLoaded, setAreFontsLoaded])

    // Fonts are loading if they are not available yet and there's no network error.
    const areFontsLoading = !(areFontsLoaded || hasNetworkError)

    return areFontsLoading
}
