import { isEqual } from 'lodash'
import { useEffect, useMemo, useState } from 'react'
import { from, merge, Subject } from 'rxjs'
import { debounceTime, distinctUntilChanged, first, map, mapTo, switchMap, throttleTime } from 'rxjs/operators'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import { dataOrThrowErrors, gql } from '../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { asError, ErrorLike, isErrorLike } from '../../../../../shared/src/util/errors'
import { queryGraphQL } from '../../../backend/graphql'
import { getCampaignExtensionData } from '../extensionData'
import { RuleDefinition } from '../../rules/types'
import { ThreadFragment } from '../../threads/util/graphql'
import { ThreadPreviewFragment } from '../preview/useCampaignPreview'
import { ActorFragment } from '../../../actor/graphql'

export const CampaignUpdatePreviewFragment = gql`
    fragment CampaignUpdatePreviewFragment on CampaignUpdatePreview {
        oldName
        newName
        oldStartDate
        newStartDate
        oldDueDate
        newDueDate
        threads {
            oldThread {
                ...ThreadFragment
            }
            newThread {
                ...ThreadPreviewFragment
            }
            operation
            oldTitle
            newTitle
        }
    }
`

const LOADING: 'loading' = 'loading'

type Result = typeof LOADING | GQL.ICampaignUpdatePreview | ErrorLike

type UpdateCampaignInputWithoutExtensionData = Pick<
    GQL.IUpdateCampaignInput,
    Exclude<keyof GQL.IUpdateCampaignInput, 'extensionData'>
>

const queryCampaignUpdatePreview = ({
    extensionsController,
    input,
}: ExtensionsControllerProps & {
    input: Pick<GQL.IUpdateCampaignInput, Exclude<keyof GQL.IUpdateCampaignInput, 'extensionData'>>
}): Promise<GQL.ICampaignUpdatePreview> =>
    getCampaignExtensionData(
        extensionsController,
        input.rules ? input.rules.map(rule => JSON.parse(rule.definition) as RuleDefinition) : []
    )
        .pipe(
            switchMap(extensionData =>
                queryGraphQL(
                    gql`
                        query CampaignUpdatePreview($input: CampaignUpdatePreviewInput!) {
                            campaignUpdatePreview(input: $input) {
                                ...CampaignUpdatePreviewFragment
                            }
                        }
                        ${CampaignUpdatePreviewFragment}
                        ${ThreadFragment}
                        ${ThreadPreviewFragment}
                        ${ActorFragment}
                    `,
                    {
                        input: {
                            campaign: input.id,
                            update: { ...input, extensionData },
                        },
                    } as GQL.ICampaignUpdatePreviewOnQueryArguments
                ).pipe(
                    map(dataOrThrowErrors),
                    map(data => data.campaignUpdatePreview)
                )
            ),
            first()
        )
        .toPromise()

/**
 * A React hook that observes a campaign update preview queried from the GraphQL API.
 *
 * TODO!(sqs): dedupe with useCampaignPreview
 */
export const useCampaignUpdatePreview = (
    { extensionsController }: ExtensionsControllerProps,
    input: UpdateCampaignInputWithoutExtensionData
): [Result, boolean] => {
    const inputSubject = useMemo(() => new Subject<UpdateCampaignInputWithoutExtensionData>(), [])
    const [isLoading, setIsLoading] = useState(true)
    const [result, setResult] = useState<Result>(LOADING)
    useEffect(() => {
        // Refresh more slowly on changes to the name or description.
        const inputSubjectChanges = merge(
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
                switchMap(input => from(queryCampaignUpdatePreview({ extensionsController, input })))
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
    useEffect(() => inputSubject.next(input), [input, inputSubject])
    return [result, isLoading]
}
