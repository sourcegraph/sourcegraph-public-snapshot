import type { ErrorLike } from '@sourcegraph/common'

import type { UsePreciseCodeIntelForPositionVariables } from '../graphql-operations'

import type { LocationsGroup } from './location'
import type { SettingsGetter } from './settings'

export interface CodeIntelData {
    references: {
        endCursor: string | null
        nodes: LocationsGroup
    }
    implementations: {
        endCursor: string | null
        nodes: LocationsGroup
    }
    prototypes: {
        endCursor: string | null
        nodes: LocationsGroup
    }
    definitions: {
        endCursor: string | null
        nodes: LocationsGroup
    }
}

export interface UseCodeIntelResult {
    data?: CodeIntelData
    error?: ErrorLike
    loading: boolean

    referencesHasNextPage: boolean
    fetchMoreReferences: () => void
    fetchMoreReferencesLoading: boolean

    implementationsHasNextPage: boolean
    fetchMoreImplementations: () => void
    fetchMoreImplementationsLoading: boolean

    prototypesHasNextPage: boolean
    fetchMorePrototypes: () => void
    fetchMorePrototypesLoading: boolean
}

export interface UseCodeIntelParameters {
    variables: UsePreciseCodeIntelForPositionVariables

    searchToken: string
    fileContent: string

    languages: string[]

    isFork: boolean
    isArchived: boolean

    getSetting: SettingsGetter
}

export type UseCodeIntel = (params: UseCodeIntelParameters) => UseCodeIntelResult
