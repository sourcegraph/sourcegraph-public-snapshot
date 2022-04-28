import { ApolloError } from '@apollo/client'

import { useMutation, useQuery } from '@sourcegraph/http-client'

import { isDefined } from '../../../../codeintel/util/helpers'
import {
    CodeIntelStatusResult,
    CodeIntelStatusVariables,
    InferedPreciseSupportLevel,
    LSIFIndexesWithRepositoryNamespaceFields,
    LsifIndexFields,
    LsifUploadFields,
    LSIFUploadsWithRepositoryNamespaceFields,
    PreciseSupportFields,
    PreciseSupportLevel,
    RequestedLanguageSupportResult,
    RequestedLanguageSupportVariables,
    RequestLanguageSupportResult,
    RequestLanguageSupportVariables,
    SearchBasedCodeIntelSupportFields,
    SearchBasedSupportLevel,
} from '../../../../graphql-operations'

import { codeIntelStatusQuery, requestedLanguageSupportQuery, requestLanguageSupportQuery } from './queries'

export interface UseCodeIntelStatusParameters {
    variables: CodeIntelStatusVariables
}

export interface UseCodeIntelStatusResult {
    data?: UseCodeIntelStatusPayload
    error?: ApolloError
    loading: boolean
}

export interface UseCodeIntelStatusPayload {
    lastIndexScan?: string
    lastUploadRetentionScan?: string
    activeUploads: LsifUploadFields[]
    recentUploads: LSIFUploadsWithRepositoryNamespaceFields[]
    recentIndexes: LSIFIndexesWithRepositoryNamespaceFields[]
    preciseSupport?: (PreciseSupportFields & { confidence?: InferedPreciseSupportLevel })[]
    searchBasedSupport?: SearchBasedCodeIntelSupportFields[]
}

export const useCodeIntelStatus = ({ variables }: UseCodeIntelStatusParameters): UseCodeIntelStatusResult => {
    const { data: rawData, error, loading } = useQuery<CodeIntelStatusResult, CodeIntelStatusVariables>(
        codeIntelStatusQuery,
        {
            variables,
            notifyOnNetworkStatusChange: false,
            fetchPolicy: 'no-cache',
        }
    )

    const repo = rawData?.repository
    const path = repo?.commit?.path
    const lsif = path?.lsif
    if (!repo || !path) {
        return { loading, error }
    }

    const summary = repo.codeIntelSummary
    const common: Omit<UseCodeIntelStatusPayload, 'preciseSupport' | 'searchBasedSupport'> = {
        lastIndexScan: summary.lastIndexScan || undefined,
        lastUploadRetentionScan: summary.lastUploadRetentionScan || undefined,
        activeUploads: lsif?.lsifUploads || [],
        recentUploads: summary.recentUploads,
        recentIndexes: summary.recentIndexes,
    }

    switch (path?.__typename) {
        case 'GitBlob': {
            const support = path.codeIntelSupport
            return {
                data: {
                    ...common,
                    searchBasedSupport: support.searchBasedSupport ? [support.searchBasedSupport] : undefined,
                    preciseSupport: support.preciseSupport ? [support.preciseSupport] : undefined,
                },
                error,
                loading,
            }
        }

        case 'GitTree': {
            const info = path.codeIntelInfo
            return {
                data: info
                    ? {
                          ...common,
                          searchBasedSupport: (info.searchBasedSupport || []).map(wrapper => wrapper.support),
                          preciseSupport: (info.preciseSupport || []).map(wrapper => ({
                              ...wrapper.support,
                              confidence: wrapper.confidence,
                          })),
                      }
                    : undefined,
                error,
                loading,
            }
        }

        default:
            return { data: undefined, error, loading }
    }
}

export interface UseRequestedLanguageSupportParameters {
    variables: RequestedLanguageSupportVariables
}

export interface UseRequestedLanguageSupportResult {
    data?: {
        languages: string[]
    }
    error?: ApolloError
    loading: boolean
}

export const useRequestedLanguageSupportQuery = ({
    variables,
}: UseRequestedLanguageSupportParameters): UseRequestedLanguageSupportResult => {
    const { data: rawData, error, loading } = useQuery<
        RequestedLanguageSupportResult,
        RequestedLanguageSupportVariables
    >(requestedLanguageSupportQuery, {
        variables,
        notifyOnNetworkStatusChange: false,
        fetchPolicy: 'no-cache',
    })

    return { data: rawData && { languages: rawData.requestedLanguageSupport }, error, loading }
}

export interface UseRequestLanguageSupportParameters {
    variables: RequestLanguageSupportVariables
    onCompleted?: () => void
}

export interface UseRequestLanguageSupportResult {
    error?: ApolloError
    loading: boolean
}

export const useRequestLanguageSupportQuery = ({
    variables,
    onCompleted,
}: UseRequestLanguageSupportParameters): [() => Promise<{}>, UseRequestLanguageSupportResult] =>
    useMutation<RequestLanguageSupportResult, RequestLanguageSupportVariables>(requestLanguageSupportQuery, {
        variables,
        onCompleted,
        notifyOnNetworkStatusChange: false,
        fetchPolicy: 'no-cache',
    })

export interface IndexerSupportMetadata {
    allIndexers: { name: string; url: string }[]
    indexerNames: string[]
    uploadsByIndexerName: Map<string, LsifUploadFields[]>
    indexesByIndexerName: Map<string, LsifIndexFields[]>
}

export function massageIndexerSupportMetadata(data: UseCodeIntelStatusPayload): IndexerSupportMetadata {
    const allUploads = data.recentUploads.flatMap(uploads => uploads.uploads)
    const uploadsByIndexerName = groupBy<LsifUploadFields, string>(allUploads, getIndexerName)
    const allIndexes = data.recentIndexes.flatMap(indexes => indexes.indexes)
    const indexesByIndexerName = groupBy<LsifIndexFields, string>(allIndexes, getIndexerName)

    const nativelySupportedIndexers = (data.preciseSupport || [])
        .filter(support => support.supportLevel === PreciseSupportLevel.NATIVE)
        .map(support => support.indexers?.[0])
        .filter(isDefined)

    const allIndexers = [
        ...groupBy(
            [...allUploads, ...allIndexes]
                .map(index => index.indexer || undefined)
                .filter(isDefined)
                .concat(nativelySupportedIndexers),
            indexer => indexer.name
        ).values(),
    ].map(indexers => indexers[0])

    const languages = [
        ...new Set(
            data.searchBasedSupport
                ?.filter(support => support.supportLevel === SearchBasedSupportLevel.BASIC)
                .map(support => support.language)
        ),
    ].sort()
    const fakeIndexerNames = languages.map(name => `lsif-${name.toLowerCase()}`)
    const indexerNames = [...new Set(allIndexers.map(indexer => indexer.name).concat(fakeIndexerNames))].sort()

    return {
        allIndexers,
        indexerNames,
        uploadsByIndexerName,
        indexesByIndexerName,
    }
}

function groupBy<V, K>(values: V[], keyFn: (value: V) => K): Map<K, V[]> {
    return values.reduce(
        (map, value) => map.set(keyFn(value), (map.get(keyFn(value)) || []).concat([value])),
        new Map<K, V[]>()
    )
}

function getIndexerName(uploadOrIndexer: LsifUploadFields | LsifIndexFields): string {
    return uploadOrIndexer.indexer?.name || ''
}
