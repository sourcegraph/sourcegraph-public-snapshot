import { BehaviorSubject, concatMap, from, map } from 'rxjs'

import { Occurrence, Range, Position, SymbolRole } from '@sourcegraph/shared/src/codeintel/scip'
import { type BlameHunkData, fetchBlameHunksMemoized } from '@sourcegraph/web/src/repo/blame/shared'
import type { CodeGraphData } from '@sourcegraph/web/src/repo/blob/codemirror/codeintel/occurrences'

import { SourcegraphURL } from '$lib/common'
import { getGraphQLClient, mapOrThrow, type GraphQLClient } from '$lib/graphql'
import { SymbolRole as GraphQLSymbolRole } from '$lib/graphql-types'
import { resolveRevision } from '$lib/repo/utils'
import { parseRepoRevision } from '$lib/shared'
import { assertNonNullable } from '$lib/utils'

import type { PageLoad, PageLoadEvent } from './$types'
import {
    BlobDiffViewCommitQuery,
    BlobFileViewBlobQuery,
    BlobFileViewCodeGraphDataQuery,
    BlobFileViewCommitQuery_revisionOverride,
    BlobFileViewHighlightedFileQuery,
} from './page.gql'

function loadDiffView({ params, url }: PageLoadEvent) {
    const client = getGraphQLClient()
    const revisionOverride = url.searchParams.get('rev')
    const { repoName } = parseRepoRevision(params.repo)
    const filePath = decodeURIComponent(params.path)

    assertNonNullable(revisionOverride, 'revisionOverride is set')

    return {
        type: 'DiffView' as const,
        enableInlineDiff: true,
        enableViewAtCommit: true,
        filePath,
        commit: client
            .query(BlobDiffViewCommitQuery, {
                repoName,
                revspec: revisionOverride,
                path: filePath,
            })
            .then(mapOrThrow(result => result.data?.repository?.commit ?? null)),
    }
}

async function fetchCodeGraphData(
    client: GraphQLClient,
    repoName: string,
    resolvedRevision: string,
    path: string
): Promise<CodeGraphData[]> {
    const response = await client.query(BlobFileViewCodeGraphDataQuery, {
        repoName,
        revspec: resolvedRevision,
        path,
    })

    const rawCodeGraphData = response.data?.repository?.commit?.blob?.codeGraphData
    if (!rawCodeGraphData) {
        return []
    }

    function translateRole(graphQLRole: GraphQLSymbolRole): SymbolRole {
        switch (graphQLRole) {
            case GraphQLSymbolRole.DEFINITION:
                return SymbolRole.Definition
            case GraphQLSymbolRole.REFERENCE:
                // TODO: is this correct?
                return SymbolRole.Import
            case GraphQLSymbolRole.FORWARD_DEFINITION:
                return SymbolRole.ForwardDefinition
            default:
                return SymbolRole.Unspecified
        }
    }

    return (
        rawCodeGraphData.map(datum => ({
            provenance: datum.provenance,
            occurrences:
                datum.occurrences?.nodes.map(
                    occ =>
                        new Occurrence(
                            new Range(
                                new Position(occ.range.start.line, occ.range.start.character),
                                new Position(occ.range.end.line, occ.range.end.character)
                            ),
                            undefined,
                            occ.symbol ?? undefined,
                            occ.roles?.map(translateRole).reduce((acc, role) => acc | role, 0)
                        )
                ) ?? [],
        })) ?? []
    )
}

async function loadFileView({ parent, params, url }: PageLoadEvent) {
    const client = getGraphQLClient()
    const revisionOverride = url.searchParams.get('rev')
    const isBlame = url.searchParams.get('view') === 'blame'
    const lineOrPosition = SourcegraphURL.from(url).lineRange
    const { repoName, revision = '' } = parseRepoRevision(params.repo)
    const resolvedRevision = revisionOverride ? Promise.resolve(revisionOverride) : resolveRevision(parent, revision)
    const filePath = decodeURIComponent(params.path)

    // Create a BehaviorSubject so preloading does not create a subscriberless observable
    const blameData = new BehaviorSubject<BlameHunkData>({ current: undefined, externalURLs: undefined })
    if (isBlame) {
        const blameHunks = from(resolvedRevision).pipe(
            concatMap(resolvedRevision => fetchBlameHunksMemoized({ repoName, revision: resolvedRevision, filePath }))
        )

        from(parent())
            .pipe(
                concatMap(({ resolvedRevision }) =>
                    blameHunks.pipe(
                        map(blameHunks => ({
                            externalURLs: resolvedRevision.repo.externalURLs,
                            current: blameHunks,
                        }))
                    )
                )
            )
            .subscribe(v => blameData.next(v))
    }

    return {
        type: 'FileView' as const,
        enableInlineDiff: true,
        enableViewAtCommit: true,
        graphQLClient: client,
        lineOrPosition,
        filePath,
        blob: resolvedRevision
            .then(resolvedRevision =>
                client.query(BlobFileViewBlobQuery, {
                    repoName,
                    revspec: resolvedRevision,
                    path: filePath,
                })
            )
            .then(mapOrThrow(result => result.data?.repository?.commit?.blob ?? null)),
        highlights: resolvedRevision
            .then(resolvedRevision =>
                client.query(BlobFileViewHighlightedFileQuery, {
                    repoName,
                    revspec: resolvedRevision,
                    path: filePath,
                    disableTimeout: false,
                })
            )
            .then(mapOrThrow(result => result.data?.repository?.commit?.blob?.highlight ?? null)),
        codeGraphData: resolvedRevision.then(resolvedRevision =>
            fetchCodeGraphData(client, repoName, resolvedRevision, filePath)
        ),
        // We can ignore the error because if the revision doesn't exist, other queries will fail as well
        revisionOverride: revisionOverride
            ? await client
                .query(BlobFileViewCommitQuery_revisionOverride, {
                    repoName,
                    revspec: revisionOverride,
                })
                .then(result => result.data?.repository?.commit)
            : null,
        externalServiceType: parent()
            .then(({ resolvedRevision }) => resolvedRevision.repo?.externalRepository?.serviceType)
            .catch(error => {
                console.error('Failed to fetch repository data:', error)
                return null
            }),
        blameData,
    }
}

export const load: PageLoad = event => {
    const showDiff = event.url.searchParams.has('diff')
    const revisionOverride = event.url.searchParams.get('rev')

    if (showDiff && revisionOverride) {
        return loadDiffView(event)
    }
    return loadFileView(event)
}
