import { useEffect, useRef, useState } from 'react'

import { fixOpenMarkdownCodeBlock } from '@sourcegraph/cody-shared/src/chat/viewHelpers'

interface UseTypewriterEffectParams {
    /**
     * Text to render using the typewriter effect.
     * You can update this gradually and the hook will adjust its rendering speed
     */
    text: string
    /**
     * Used to control the speed of the typewriter effect.
     * The delay between characters is calculated by dividing `baseDelay` by the number of remaining characters to render,
     * This means that when there are more characters left to render, the delay is reduced, making the typing effect faster.
     * If the number of remaining characters is small, the delay is larger, slowing down the typing effect.
     * This behavior is designed to keep the rendering as close as possible to the most current state of the string being rendered.
     * Note: There is a minimum delay limit of 10ms to prevent the typing speed from being too fast to perceive, and to avoid potential performance issues.
     */
    baseDelay: number
    /**
     * Used to render text immediately, skipping the typewriter effect.
     * This should be used for any text that is known ahead of time, or if we want to end the effect early (i.e. we know no new text will come).
     */
    immediate?: boolean
}

export function useTypewriterEffect({ text, baseDelay, immediate }: UseTypewriterEffectParams): string {
    const [renderedText, setRenderedText] = useState(immediate ? text : '')
    const currentText = useRef('')

    useEffect(() => {
        let intervalId: number | null = null

        if (!immediate && currentText.current.length < text.length) {
            const charactersLeft = text.length - currentText.current.length
            const dynamicDelay = Math.max(baseDelay / charactersLeft, 10) // set a minimum delay of 5ms
            intervalId = window.setInterval(() => {
                currentText.current = text.slice(0, currentText.current.length + 1)
                setRenderedText(currentText.current)
            }, dynamicDelay)
        }

        return () => {
            if (intervalId !== null) {
                window.clearInterval(intervalId)
            }
        }
    }, [immediate, text, baseDelay])

    return fixOpenMarkdownCodeBlock(renderedText)
}
