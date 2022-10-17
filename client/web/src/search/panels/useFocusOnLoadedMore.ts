import { useCallback, useEffect, useRef, useState } from 'react'

import { useCallbackRef } from 'use-callback-ref'

export const useFocusOnLoadedMore = (
    numberOfProcessedResults: number
): ((index: number) => React.MutableRefObject<HTMLAnchorElement | null> | undefined) => {
    const numberOfItems = useRef<number>(numberOfProcessedResults)
    const [firstLoadedMoreItemIndex, setFirstLoadedMoreItemIndex] = useState<number>(-1)
    const firstLoadedMoreItemCallbackRef = useCallbackRef<HTMLAnchorElement | null>(null, nextRef => nextRef?.focus())

    useEffect(() => {
        if (numberOfItems.current !== 0 && numberOfItems.current !== numberOfProcessedResults) {
            setFirstLoadedMoreItemIndex(numberOfItems.current)
        }

        numberOfItems.current = numberOfProcessedResults
    }, [numberOfProcessedResults])

    const getItemRef = useCallback(
        (index: number) => (index === firstLoadedMoreItemIndex ? firstLoadedMoreItemCallbackRef : undefined),
        [firstLoadedMoreItemIndex, firstLoadedMoreItemCallbackRef]
    )

    return getItemRef
}
