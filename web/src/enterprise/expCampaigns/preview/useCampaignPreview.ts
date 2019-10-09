import { isEqual } from 'lodash'
import { useEffect, useMemo, useState } from 'react'
import { combineLatest, merge, Observable, of, Subject } from 'rxjs'
import {
    catchError,
    debounceTime,
    distinctUntilChanged,
    map,
    mapTo,
    switchMap,
    tap,
    throttleTime,
} from 'rxjs/operators'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import { dataOrThrowErrors, gql } from '../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { asError, ErrorLike, isErrorLike } from '../../../../../shared/src/util/errors'
import { ActorFragment, ActorQuery } from '../../../actor/graphql'
import { queryGraphQL } from '../../../backend/graphql'
import {
    diffStatFieldsFragment,
    fileDiffHunkRangeFieldsFragment,
} from '../../../repo/compare/RepositoryCompareDiffPage'
import { RuleDefinition } from '../../rules/types'
import { ThreadConnectionFiltersFragment } from '../../threads/list/useThreads'
import { ExtensionDataStatus, getCampaignExtensionData } from '../extensionData'

export const RepositoryComparisonQuery = gql`
baseRepository {
    id
    name
    url
}
headRepository {
    id
    name
    url
}
range {
    expr
    baseRevSpec {
        object {
            oid
        }
        expr
    }
    headRevSpec {
        expr
    }
}
fileDiffs {
    nodes {
        oldPath
        newPath
        hunks {
            oldRange {
                ...FileDiffHunkRangeFields
            }
            oldNoNewlineAt
            newRange {
                ...FileDiffHunkRangeFields
            }
            section
            body
        }
        stat {
            ...DiffStatFields
        }
        internalID
    }
    totalCount
    pageInfo {
        hasNextPage
    }
    diffStat {
        ...DiffStatFields
    }
}`

export const ThreadPreviewFragment = gql`
fragment ThreadPreviewFragment on ThreadPreview {
    author {
        ${ActorQuery}
    }
    repository {
        id
        name
        url
    }
    title
    bodyHTML
    isDraft
    isPendingExternalCreation
    kind
    assignees {
        nodes {
            ${ActorQuery}
        }
    }
    internalID
}`

export const CampaignPreviewFragment = gql`
    fragment CampaignPreviewFragment on ExpCampaignPreview {
        name
        threads(filters: { query: $query }) {
            nodes {
                __typename
                ... on ThreadPreview {
                    ...ThreadPreviewFragment
                    repositoryComparison {
                        ${RepositoryComparisonQuery}
                    }
                }
            }
            totalCount
            filters {
                ...ThreadConnectionFiltersFragment
            }
        }
        participants {
            edges {
                actor {
                    ${ActorQuery}
                }
                reasons
            }
            totalCount
        }
        diagnostics {
            nodes {
                type
                data
            }
            totalCount
        }
        repositories {
            id
        }
        repositoryComparisons {
            ${RepositoryComparisonQuery}
        }
    }
    ${ThreadPreviewFragment}
    ${fileDiffHunkRangeFieldsFragment}
    ${diffStatFieldsFragment}
`

const LOADING: 'loading' = 'loading'

type Result = typeof LOADING | GQL.IExpCampaignPreview | ErrorLike

type CreateCampaignInputWithoutExtensionData = Pick<
    GQL.IExpCreateCampaignInput,
    Exclude<keyof GQL.IExpCreateCampaignInput, 'extensionData'>
>

const queryCampaignPreview = ({
    extensionsController,
    input,
    query,
}: ExtensionsControllerProps & {
    input: Pick<GQL.IExpCreateCampaignInput, Exclude<keyof GQL.IExpCreateCampaignInput, 'extensionData'>>
    query: string
}): Observable<[GQL.IExpCampaignPreview | ErrorLike, ExtensionDataStatus]> => {
    const extensionDataAndStatus = getCampaignExtensionData(
        extensionsController,
        input.rules ? input.rules.map(rule => JSON.parse(rule.definition) as RuleDefinition) : []
    )
    const campaignPreview = extensionDataAndStatus.pipe(
        map(([extensionData]) => extensionData),
        switchMap(extensionData =>
            queryGraphQL(
                gql`
                    query CampaignPreview($input: ExpCampaignPreviewInput!, $query: String) {
                        expCampaignPreview(input: $input) {
                            ...CampaignPreviewFragment
                        }
                    }
                    ${CampaignPreviewFragment}
                    ${ThreadConnectionFiltersFragment}
                    ${ActorFragment}
                `,
                // tslint:disable-next-line: no-object-literal-type-assertion
                {
                    input: {
                        campaign: {
                            ...input,
                            extensionData,
                        },
                    },
                    query,
                } as GQL.IExpCampaignPreviewOnQueryArguments
            ).pipe(
                map(dataOrThrowErrors),
                map(data => data.expCampaignPreview),
                tap(data => {
                    // TODO!(sqs) hack, compensate for the RepositoryComparison head not existing
                    const fixup = (c: GQL.IRepositoryComparison): void => {
                        if (c.range) {
                            c.range.headRevSpec.object = { oid: '' } as any
                        }
                        for (const d of c.fileDiffs.nodes) {
                            d.mostRelevantFile = { path: d.newPath, url: '' } as any
                        }
                    }
                    for (const c of data.repositoryComparisons) {
                        fixup(c)
                    }
                    for (const t of data.threads.nodes) {
                        if (t.repositoryComparison) {
                            fixup(t.repositoryComparison)
                        }
                    }
                }),
                catchError(err => of(asError(err)))
            )
        )
    )
    return combineLatest([campaignPreview, extensionDataAndStatus.pipe(map(([, status]) => status))])
}

/**
 * A React hook that observes a campaign preview queried from the GraphQL API.
 */
export const useCampaignPreview = (
    { extensionsController }: ExtensionsControllerProps,
    input: CreateCampaignInputWithoutExtensionData,
    query: string // TODO!(sqs): the query param is currently ignored
): [Result, ExtensionDataStatus, boolean] => {
    const inputSubject = useMemo(() => new Subject<CreateCampaignInputWithoutExtensionData>(), [])
    const [isLoading, setIsLoading] = useState(true)
    const [result, setResult] = useState<Result>(LOADING)
    const [status, setStatus] = useState<ExtensionDataStatus>({ message: '' })
    useEffect(() => {
        // Refresh more slowly on changes to the name or description.
        const inputSubjectChanges = merge(
            inputSubject.pipe(distinctUntilChanged((a, b) => a.namespace === b.namespace && a.draft === b.draft)),
            inputSubject.pipe(
                debounceTime(250),
                distinctUntilChanged((a, b) => isEqual(a.rules, b.rules))
            ),
            inputSubject.pipe(
                distinctUntilChanged(
                    (a, b) =>
                        a.name === b.name && a.body === b.body && a.startDate === b.startDate && a.dueDate === b.dueDate
                ),
                debounceTime(2000)
            )
        )
        const subscription = merge(
            inputSubjectChanges.pipe(
                distinctUntilChanged((a, b) => isEqual(a, b)),
                mapTo([LOADING, { message: 'LOADING123' }] as [typeof LOADING, ExtensionDataStatus])
            ),
            inputSubjectChanges.pipe(
                throttleTime(1000, undefined, { leading: true, trailing: true }),
                distinctUntilChanged((a, b) => isEqual(a, b)),
                switchMap(input => queryCampaignPreview({ extensionsController, input, query }))
            )
        ).subscribe(([result, status]) => {
            setStatus(status)
            setResult(prevResult => {
                setIsLoading(result === LOADING)
                // Reuse last result while loading, to reduce UI jitter.
                return result === LOADING && prevResult !== LOADING
                    ? isErrorLike(prevResult)
                        ? LOADING
                        : prevResult
                    : result
            })
        })
        return () => subscription.unsubscribe()
    }, [extensionsController, inputSubject, query])
    useEffect(() => inputSubject.next(input), [input, inputSubject])
    return [result, status, isLoading]
}
