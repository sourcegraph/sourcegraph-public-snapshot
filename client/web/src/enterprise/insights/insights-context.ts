import { createContext } from 'use-context-selector'

import { SearchBoxProps } from '../../search/input/SearchBox'

interface InsightsContext {
    searchBoxProps: Omit<SearchBoxProps, 'queryState' | 'onChange' | 'onSubmit' | 'isSearchOnboardingTourVisible'>
}

export const InsightsContext = createContext<InsightsContext | undefined>(undefined)
