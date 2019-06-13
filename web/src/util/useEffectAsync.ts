import { EffectCallback, useEffect } from 'react'

/**
 * TODO!(sqs): This function is bad practice. See
 * https://github.com/facebook/react/issues/14326#issuecomment-472043812 and
 * https://www.robinwieruch.de/react-hooks-fetch-data/.
 */
export function useEffectAsync(
    effect: EffectCallback | (() => Promise<void>),
    inputs: readonly any[] | undefined
): void {
    useEffect(() => {
        try {
            effect()
        } catch (err) {
            console.error('useEffectAsync:', err)
            throw err
        }
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, inputs)
}
