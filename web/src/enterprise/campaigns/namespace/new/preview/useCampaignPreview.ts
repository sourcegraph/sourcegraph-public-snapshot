import { useEffect, useMemo, useState } from 'react'
import { merge, Subject } from 'rxjs'
import { map, mapTo, switchMap, tap, throttleTime } from 'rxjs/operators'
import { ExtensionsControllerProps } from '../../../../../../../shared/src/extensions/controller'
import { dataOrThrowErrors, gql } from '../../../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../../../shared/src/graphql/schema'
import { asError, ErrorLike, isErrorLike } from '../../../../../../../shared/src/util/errors'
import { queryGraphQL } from '../../../../../backend/graphql'
import {
    diffStatFieldsFragment,
    fileDiffHunkRangeFieldsFragment,
} from '../../../../../repo/compare/RepositoryCompareDiffPage'
import { getCampaignExtensionData } from '../../../extensionData'

export const CampaignPreviewFragment = gql`
    fragment CampaignPreviewFragment on CampaignPreview {
        name
        threads {
            nodes {
                __typename
                ... on ThreadPreview {
                    repository {
                        id
                        name
                        url
                    }
                    title
                    kind
                }
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
            }
        }
    }
    ${fileDiffHunkRangeFieldsFragment}
    ${diffStatFieldsFragment}
`

const LOADING: 'loading' = 'loading'

type Result = typeof LOADING | GQL.ICampaignPreview | ErrorLike

type CreateCampaignInputWithoutExtensionData = Pick<
    GQL.ICreateCampaignInput,
    Exclude<keyof GQL.ICreateCampaignInput, 'extensionData'>
>

/**
 * A React hook that observes a campaign preview queried from the GraphQL API.
 */
export const useCampaignPreview = (
    { extensionsController }: ExtensionsControllerProps,
    input: CreateCampaignInputWithoutExtensionData
): [Result, boolean] => {
    const inputSubject = useMemo(() => new Subject<CreateCampaignInputWithoutExtensionData>(), [])
    const [isLoading, setIsLoading] = useState(true)
    const [result, setResult] = useState<Result>(LOADING)
    useEffect(() => inputSubject.next(input), [input, inputSubject])
    useEffect(() => {
        const subscription = merge(
            inputSubject.pipe(mapTo(LOADING)),
            inputSubject.pipe(
                throttleTime(1000, undefined, { leading: true, trailing: true }),
                switchMap(input =>
                    getCampaignExtensionData(extensionsController, input).pipe(
                        switchMap(extensionData =>
                            queryGraphQL(
                                gql`
                                    query CampaignPreview($input: CampaignPreviewInput!) {
                                        campaignPreview(input: $input) {
                                            ...CampaignPreviewFragment
                                        }
                                    }
                                    ${CampaignPreviewFragment}
                                `,
                                // tslint:disable-next-line: no-object-literal-type-assertion
                                {
                                    input: {
                                        campaign: {
                                            ...input,
                                            extensionData,
                                        },
                                    },
                                } as GQL.ICampaignPreviewOnQueryArguments
                            ).pipe(
                                map(dataOrThrowErrors),
                                map(data => data.campaignPreview),
                                tap(data => {
                                    // TODO!(sqs) hack, compensate for the RepositoryComparison head not existing
                                    for (const c of data.repositoryComparisons) {
                                        c.range.headRevSpec.object = { oid: '' } as any
                                        for (const d of c.fileDiffs.nodes) {
                                            d.mostRelevantFile = { path: d.newPath, url: '' } as any
                                        }
                                    }
                                })
                            )
                        )
                    )
                )
            )
        ).subscribe(
            result => {
                setResult(prevResult => {
                    setIsLoading(result === LOADING)
                    // Reuse last non-error result while loading, to reduce UI jitter.
                    return result === LOADING && prevResult !== LOADING && !isErrorLike(prevResult)
                        ? prevResult
                        : result
                })
            },
            err => {
                setIsLoading(false)
                setResult(asError(err))
            }
        )
        return () => subscription.unsubscribe()
    }, [extensionsController, inputSubject])
    return [result, isLoading]
}
