import { useEffect, useRef, useState } from 'react'

import type { QueryResult } from '@apollo/client'

import { dataOrThrowErrors, useLazyQuery, useQuery } from '@sourcegraph/http-client'

import { type Location, buildPreciseLocation, LocationsGroup } from '../../codeintel/location'
import {
    LOAD_ADDITIONAL_IMPLEMENTATIONS_QUERY,
    LOAD_ADDITIONAL_PROTOTYPES_QUERY,
    LOAD_ADDITIONAL_REFERENCES_QUERY,
    USE_PRECISE_CODE_INTEL_FOR_POSITION_QUERY,
} from '../../codeintel/ReferencesPanelQueries'
import type { CodeIntelData, UseCodeIntelParameters, UseCodeIntelResult } from '../../codeintel/useCodeIntel'
import type { ConnectionQueryArguments } from '../../components/FilteredConnection'
import { asGraphQLResult } from '../../components/FilteredConnection/utils'
import type {
    UsePreciseCodeIntelForPositionVariables,
    UsePreciseCodeIntelForPositionResult,
    LoadAdditionalReferencesResult,
    LoadAdditionalReferencesVariables,
    LoadAdditionalImplementationsResult,
    LoadAdditionalImplementationsVariables,
    LoadAdditionalPrototypesResult,
    LoadAdditionalPrototypesVariables,
} from '../../graphql-operations'

import { useSearchBasedCodeIntel } from './useSearchBasedCodeIntel'

const EMPTY_CODE_INTEL_DATA: CodeIntelData = {
    implementations: { endCursor: null, nodes: LocationsGroup.empty },
    prototypes: { endCursor: null, nodes: LocationsGroup.empty },
    definitions: { endCursor: null, nodes: LocationsGroup.empty },
    references: { endCursor: null, nodes: LocationsGroup.empty },
}

export const useCodeIntel = ({
    variables,
    searchToken,
    languages,
    fileContent,
    isFork,
    isArchived,
    getSetting,
}: UseCodeIntelParameters): UseCodeIntelResult => {
    const shouldMixPreciseAndSearchBasedReferences = (): boolean =>
        getSetting<boolean>('codeIntel.mixPreciseAndSearchBasedReferences', false)

    const [codeIntelData, setCodeIntelData] = useState<CodeIntelData>()

    const fellBackToSearchBased = useRef(false)
    const shouldFetchPrecise = useRef(true)
    useEffect(() => {
        // We need to fetch again if the variables change
        shouldFetchPrecise.current = true
        fellBackToSearchBased.current = false
    }, [
        variables.repository,
        variables.commit,
        variables.path,
        variables.line,
        variables.character,
        variables.filter,
        variables.firstReferences,
        variables.firstImplementations,
    ])

    const {
        loading: searchBasedLoading,
        error: searchBasedError,
        fetch: fetchSearchBasedCodeIntel,
        fetchReferences: fetchSearchBasedReferences,
        fetchDefinitions: fetchSearchBasedDefinitions,
    } = useSearchBasedCodeIntel({
        repo: variables.repository,
        commit: variables.commit,
        path: variables.path,
        filter: variables.filter ?? undefined,
        searchToken,
        position: {
            line: variables.line,
            character: variables.character,
        },
        fileContent,
        languages,
        isFork,
        isArchived,
        getSetting,
    })

    const { error, loading } = useQuery<
        UsePreciseCodeIntelForPositionResult,
        UsePreciseCodeIntelForPositionVariables & ConnectionQueryArguments
    >(USE_PRECISE_CODE_INTEL_FOR_POSITION_QUERY, {
        variables,
        notifyOnNetworkStatusChange: false,
        fetchPolicy: 'no-cache',
        onCompleted: result => {
            if (!shouldFetchPrecise.current) {
                return
            }
            shouldFetchPrecise.current = false

            let refs: CodeIntelData['references'] = { endCursor: null, nodes: LocationsGroup.empty }
            let defs: CodeIntelData['definitions'] = { endCursor: null, nodes: LocationsGroup.empty }
            const addRefs = (newRefs: Location[]): void => {
                refs.nodes = refs.nodes.combine(newRefs)
            }
            const addDefs = (newDefs: Location[]): void => {
                defs.nodes = defs.nodes.combine(newDefs)
            }

            const lsifData = result ? getLsifData({ data: result }) : undefined
            if (lsifData) {
                refs = lsifData.references
                defs = lsifData.definitions
                // If we've exhausted LSIF data and the flag is enabled, we add search-based data.
                if (refs.endCursor === null && shouldMixPreciseAndSearchBasedReferences()) {
                    fetchSearchBasedReferences(addRefs)
                }
                // When no definitions are found, the hover tooltip falls back to a search based
                // search, regardless of the mixPreciseAndSearchBasedReferences setting.
                if (defs.nodes.locationsCount === 0) {
                    fetchSearchBasedDefinitions(addDefs)
                }
            } else {
                fellBackToSearchBased.current = true
                fetchSearchBasedCodeIntel(addRefs, addDefs)
            }
            setCodeIntelData({
                ...(lsifData || EMPTY_CODE_INTEL_DATA),
                definitions: defs,
                references: refs,
            })
        },
    })

    const [fetchAdditionalReferences, additionalReferencesResult] = useLazyQuery<
        LoadAdditionalReferencesResult,
        LoadAdditionalReferencesVariables & ConnectionQueryArguments
    >(LOAD_ADDITIONAL_REFERENCES_QUERY, {
        fetchPolicy: 'no-cache',
        onCompleted: result => {
            const previousData = codeIntelData
            const newReferenceData = result.repository?.commit?.blob?.lsif?.references
            if (!previousData || !newReferenceData) {
                return
            }
            let references: LocationsGroup = previousData.references.nodes.combine(
                newReferenceData.nodes.map(buildPreciseLocation)
            )
            const endCursor = newReferenceData.pageInfo.endCursor
            if (endCursor === null && shouldMixPreciseAndSearchBasedReferences()) {
                // If we've exhausted LSIF data and the flag is enabled, we add search-based data.
                fetchSearchBasedReferences((refs: Location[]) => {
                    references = references.combine(refs)
                })
            }
            setCodeIntelData({
                ...previousData,
                references: {
                    endCursor,
                    nodes: references,
                },
            })
        },
    })

    const [fetchAdditionalPrototypes, additionalPrototypesResult] = useLazyQuery<
        LoadAdditionalPrototypesResult,
        LoadAdditionalPrototypesVariables & ConnectionQueryArguments
    >(LOAD_ADDITIONAL_PROTOTYPES_QUERY, {
        fetchPolicy: 'no-cache',
        onCompleted: result => {
            const previousData = codeIntelData
            const newPrototypesData = result.repository?.commit?.blob?.lsif?.prototypes
            if (!previousData || !newPrototypesData) {
                return
            }
            setCodeIntelData({
                ...previousData,
                prototypes: {
                    endCursor: newPrototypesData.pageInfo.endCursor,
                    nodes: previousData.prototypes.nodes.combine(newPrototypesData.nodes.map(buildPreciseLocation)),
                },
            })
        },
    })

    const [fetchAdditionalImplementations, additionalImplementationsResult] = useLazyQuery<
        LoadAdditionalImplementationsResult,
        LoadAdditionalImplementationsVariables & ConnectionQueryArguments
    >(LOAD_ADDITIONAL_IMPLEMENTATIONS_QUERY, {
        fetchPolicy: 'no-cache',
        onCompleted: result => {
            const previousData = codeIntelData
            const newImplementationsData = result.repository?.commit?.blob?.lsif?.implementations
            if (!previousData || !newImplementationsData) {
                return
            }
            setCodeIntelData({
                ...previousData,
                implementations: {
                    endCursor: newImplementationsData.pageInfo.endCursor,
                    nodes: previousData.implementations.nodes.combine(
                        newImplementationsData.nodes.map(buildPreciseLocation)
                    ),
                },
            })
        },
    })

    const fetchMoreReferences = (): void => {
        const cursor = codeIntelData?.references.endCursor || null

        if (cursor !== null) {
            // eslint-disable-next-line @typescript-eslint/no-floating-promises
            fetchAdditionalReferences({
                variables: {
                    ...variables,
                    ...{ afterReferences: cursor },
                },
            })
        }
    }

    const fetchMoreImplementations = (): void => {
        const cursor = codeIntelData?.implementations.endCursor || null

        if (cursor !== null) {
            // eslint-disable-next-line @typescript-eslint/no-floating-promises
            fetchAdditionalImplementations({
                variables: {
                    ...variables,
                    ...{ afterImplementations: cursor },
                },
            })
        }
    }

    const fetchMorePrototypes = (): void => {
        const cursor = codeIntelData?.prototypes.endCursor || null

        if (cursor !== null) {
            // eslint-disable-next-line @typescript-eslint/no-floating-promises
            fetchAdditionalPrototypes({
                variables: {
                    ...variables,
                    ...{ afterPrototypes: cursor },
                },
            })
        }
    }

    const combinedLoading = loading || (fellBackToSearchBased.current && searchBasedLoading)

    const combinedError = error || searchBasedError

    return {
        data: codeIntelData,
        loading: combinedLoading,

        error: combinedError,

        fetchMoreReferences,
        fetchMoreReferencesLoading: additionalReferencesResult.loading,
        referencesHasNextPage: codeIntelData ? codeIntelData.references.endCursor !== null : false,

        fetchMoreImplementations,
        implementationsHasNextPage: codeIntelData ? codeIntelData.implementations.endCursor !== null : false,
        fetchMoreImplementationsLoading: additionalImplementationsResult.loading,

        fetchMorePrototypes,
        prototypesHasNextPage: codeIntelData ? codeIntelData.prototypes.endCursor !== null : false,
        fetchMorePrototypesLoading: additionalPrototypesResult.loading,
    }
}

const getLsifData = ({
    data,
    error,
}: Pick<QueryResult<UsePreciseCodeIntelForPositionResult>, 'data' | 'error'>): CodeIntelData | undefined => {
    const result = asGraphQLResult({ data, errors: error?.graphQLErrors || [] })

    const extractedData = dataOrThrowErrors(result)

    // If there weren't any errors and we just didn't receive any data
    if (!extractedData?.repository?.commit?.blob?.lsif) {
        return undefined
    }

    const lsif = extractedData.repository?.commit?.blob?.lsif

    return {
        implementations: {
            endCursor: lsif.implementations.pageInfo.endCursor,
            nodes: new LocationsGroup(lsif.implementations.nodes.map(buildPreciseLocation)),
        },
        prototypes: {
            endCursor: lsif.prototypes.pageInfo.endCursor,
            nodes: new LocationsGroup(lsif.prototypes.nodes.map(buildPreciseLocation)),
        },
        references: {
            endCursor: lsif.references.pageInfo.endCursor,
            nodes: new LocationsGroup(lsif.references.nodes.map(buildPreciseLocation)),
        },
        definitions: {
            endCursor: lsif.definitions.pageInfo.endCursor,
            nodes: new LocationsGroup(lsif.definitions.nodes.map(buildPreciseLocation)),
        },
    }
}
