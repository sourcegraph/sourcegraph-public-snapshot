import { isEqual } from 'lodash'
import { useEffect, useMemo, useState } from 'react'
import { from, merge, Subject } from 'rxjs'
import { debounceTime, distinctUntilChanged, first, map, mapTo, switchMap, tap, throttleTime } from 'rxjs/operators'
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
import { ThreadConnectionFiltersFragment } from '../../threads/list/useThreads'
import { getCampaignExtensionData } from '../extensionData'
import { RuleDefinition } from '../../rules/types'

const RepositoryComparisonQuery = gql`
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
    kind
    assignees {
        nodes {
            ${ActorQuery}
        }
    }
    internalID
}`

export const CampaignPreviewFragment = gql`
    fragment CampaignPreviewFragment on CampaignPreview {
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

type Result = typeof LOADING | GQL.ICampaignPreview | ErrorLike

type CreateCampaignInputWithoutExtensionData = Pick<
    GQL.ICreateCampaignInput,
    Exclude<keyof GQL.ICreateCampaignInput, 'extensionData'>
>

const queryCampaignPreview = ({
    extensionsController,
    input,
    query,
}: ExtensionsControllerProps & {
    input: Pick<GQL.ICreateCampaignInput, Exclude<keyof GQL.ICreateCampaignInput, 'extensionData'>>
    query: string
}): Promise<GQL.ICampaignPreview> =>
    getCampaignExtensionData(
        extensionsController,
        input.rules ? input.rules.map(rule => JSON.parse(rule.definition) as RuleDefinition) : []
    )
        .pipe(
            switchMap(extensionData =>
                queryGraphQL(
                    gql`
                        query CampaignPreview($input: CampaignPreviewInput!, $query: String) {
                            campaignPreview(input: $input) {
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
                    } as GQL.ICampaignPreviewOnQueryArguments
                ).pipe(
                    map(dataOrThrowErrors),
                    map(data => data.campaignPreview),
                    tap(data => {
                        // TODO!(sqs) hack, compensate for the RepositoryComparison head not existing
                        const fixup = (c: GQL.IRepositoryComparison): void => {
                            c.range.headRevSpec.object = { oid: '' } as any
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
                    })
                )
            ),
            first()
        )
        .toPromise()

/**
 * A React hook that observes a campaign preview queried from the GraphQL API.
 */
export const useCampaignPreview = (
    { extensionsController }: ExtensionsControllerProps,
    input: CreateCampaignInputWithoutExtensionData,
    query: string // TODO!(sqs): the query param is currently ignored
): [Result, boolean] => {
    const inputSubject = useMemo(() => new Subject<CreateCampaignInputWithoutExtensionData>(), [])
    const [isLoading, setIsLoading] = useState(true)
    const [result, setResult] = useState<Result>(LOADING)
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
            inputSubjectChanges.pipe(mapTo(LOADING)),
            inputSubjectChanges.pipe(
                throttleTime(1000, undefined, { leading: true, trailing: true }),
                switchMap(input => from(queryCampaignPreview({ extensionsController, input, query })))
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
    }, [extensionsController, inputSubject, query])
    useEffect(() => inputSubject.next(input), [input, inputSubject])
    return [result, isLoading]
}
