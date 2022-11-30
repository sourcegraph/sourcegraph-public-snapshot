import { useEffect, useState } from 'react'

import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'

import type { FuzzyFinderContainerProps } from './FuzzyFinder'

const FuzzyFinderContainer = lazyComponent(() => import('./FuzzyFinder'), 'FuzzyFinderContainer')

export const LazyFuzzyFinder: React.FunctionComponent<FuzzyFinderContainerProps> = props => {
    const [isLoaded, setIsLoaded] = useState(false)
    useEffect(() => {
        // It's OK to always load the fuzzy finder, we just want to delay until
        // after the initial page render. We delay the request by a small number
        // of milliseconds to allow other useEffect-triggered GQL request to
        // take precedence.
        const id = setTimeout(() => setIsLoaded(true), 100)
        return () => clearTimeout(id)
    }, [])
    return isLoaded ? <FuzzyFinderContainer {...props} /> : null
}
