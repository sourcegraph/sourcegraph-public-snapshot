import { useEffect, useMemo, useState } from 'react'
import { merge, Observable, Subject, of, combineLatest } from 'rxjs'
import {
    catchError,
    debounceTime,
    distinctUntilChanged,
    map,
    startWith,
    switchMap,
    throttleTime,
    share,
} from 'rxjs/operators'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import { dataOrThrowErrors, gql } from '../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { asError, ErrorLike, isErrorLike } from '../../../../../shared/src/util/errors'
import { ActorFragment } from '../../../actor/graphql'
import { queryGraphQL } from '../../../backend/graphql'
import {
    diffStatFieldsFragment,
    fileDiffHunkRangeFieldsFragment,
} from '../../../repo/compare/RepositoryCompareDiffPage'
import { ThreadFragment } from '../../threads/util/graphql'
import { getCampaignExtensionData, ExtensionDataStatus } from '../extensionData'
import { RepositoryComparisonQuery, ThreadPreviewFragment, EMPTY_EXT_DATA } from '../preview/useCampaignPreview'
import { parseJSONCOrError } from '../../../../../shared/src/util/jsonc'
import { Workflow } from '../../../schema/workflow.schema'

export const CampaignUpdatePreviewFragment = gql`
    fragment CampaignUpdatePreviewFragment on ExpCampaignUpdatePreview {
        oldName
        newName
        oldStartDate
        newStartDate
        oldDueDate
        newDueDate
        threads {
            __typename
            oldThread {
                ...ThreadFragment
                repositoryComparison {
                    ${RepositoryComparisonQuery}
                }
            }
            newThread {
                ...ThreadPreviewFragment
                repositoryComparison {
                    ${RepositoryComparisonQuery}
                }
            }
            operation
            oldTitle
            newTitle
        }
        repositoryComparisons {
            old {
                ${RepositoryComparisonQuery}
            }
            new {
                ${RepositoryComparisonQuery}
            }
        }
    }
`

const LOADING: 'loading' = 'loading'

type Result = typeof LOADING | GQL.IExpCampaignUpdatePreview | ErrorLike

type UpdateCampaignInputWithoutExtensionData = Pick<
    GQL.IExpUpdateCampaignInput,
    Exclude<keyof GQL.IExpUpdateCampaignInput, 'extensionData'>
>

const queryCampaignUpdatePreview = ({
    extensionsController,
    input,
}: ExtensionsControllerProps & {
    input: Pick<GQL.IExpUpdateCampaignInput, Exclude<keyof GQL.IExpUpdateCampaignInput, 'extensionData'>>
}): Observable<
    readonly [GQL.IExpCampaignUpdatePreview | ErrorLike, GQL.IExpCampaignExtensionData, ExtensionDataStatus]
> => {
    const workflow = parseJSONCOrError<Workflow>(input.workflowAsJSONCString)
    if (isErrorLike(workflow)) {
        return of([workflow, EMPTY_EXT_DATA, { isLoading: false }])
    }

    const extensionDataAndStatus = getCampaignExtensionData(extensionsController, workflow, input).pipe(share())
    const campaignUpdatePreview = extensionDataAndStatus.pipe(
        switchMap(extensionData =>
            queryGraphQL(
                gql`
                    query CampaignUpdatePreview($input: ExpCampaignUpdatePreviewInput!) {
                        expCampaignUpdatePreview(input: $input) {
                            ...CampaignUpdatePreviewFragment
                        }
                    }
                    ${CampaignUpdatePreviewFragment}
                    ${ThreadFragment}
                    ${ThreadPreviewFragment}
                    ${ActorFragment}
                    ${fileDiffHunkRangeFieldsFragment}
                    ${diffStatFieldsFragment}
                `,
                {
                    input: {
                        campaign: input.id,
                        update: { ...input, extensionData },
                    },
                } as GQL.IExpCampaignUpdatePreviewOnQueryArguments
            ).pipe(
                map(dataOrThrowErrors),
                map(data => data.expCampaignUpdatePreview)
            )
        )
    )
    return combineLatest([campaignUpdatePreview, extensionDataAndStatus]).pipe(
        map(([campaignUpdatePreview, [data, status]]) => [campaignUpdatePreview, data, status] as const)
    )
}

/**
 * A React hook that observes a campaign update preview queried from the GraphQL API.
 *
 * TODO!(sqs): dedupe with useCampaignPreview
 */
export const useCampaignUpdatePreview = (
    { extensionsController }: ExtensionsControllerProps,
    input: UpdateCampaignInputWithoutExtensionData
): [Result | null, GQL.IExpCampaignExtensionData, boolean] => {
    const inputSubject = useMemo(() => new Subject<UpdateCampaignInputWithoutExtensionData>(), [])
    const [isLoading, setIsLoading] = useState(true)
    const [result, setResult] = useState<Result>(LOADING)
    const [data] = useState<GQL.IExpCampaignExtensionData>(EMPTY_EXT_DATA)
    useEffect(() => {
        // Refresh more slowly on changes to the name or description.
        const inputSubjectChanges = merge(
            inputSubject.pipe(
                debounceTime(250),
                distinctUntilChanged((a, b) => a.workflowAsJSONCString === b.workflowAsJSONCString)
            ),
            inputSubject.pipe(
                distinctUntilChanged(
                    (a, b) =>
                        a.name === b.name && a.body === b.body && a.startDate === b.startDate && a.dueDate === b.dueDate
                ),
                debounceTime(2000)
            )
        )
        const subscription = inputSubjectChanges
            .pipe(
                throttleTime(1000, undefined, { leading: true, trailing: true }),
                switchMap(input =>
                    queryCampaignUpdatePreview({ extensionsController, input }).pipe(catchError(err => [asError(err)]))
                ),
                startWith(LOADING)
            )
            .subscribe(resultOrError => {
                if (isErrorLike(resultOrError)) {
                    setIsLoading(false)
                    setResult(resultOrError)
                    return
                }
                setResult(prevResult => {
                    setIsLoading(resultOrError === LOADING)
                    // Reuse last non-error result while loading, to reduce UI jitter.
                    return resultOrError === LOADING && prevResult !== LOADING && !isErrorLike(prevResult)
                        ? prevResult
                        : resultOrError
                })
            })
        return () => subscription.unsubscribe()
    }, [extensionsController, inputSubject])
    useEffect(() => inputSubject.next(input), [input, inputSubject])
    return [result, data, isLoading]
}
