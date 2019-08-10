import { useEffect, useState } from 'react'
import { combineLatest, from } from 'rxjs'
import { map, startWith, switchMap } from 'rxjs/operators'
import { fromDiagnostic } from '../../../../../../../shared/src/api/types/diagnostic'
import { ExtensionsControllerProps } from '../../../../../../../shared/src/extensions/controller'
import { dataOrThrowErrors, gql } from '../../../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../../../shared/src/graphql/schema'
import { asError, ErrorLike, isErrorLike } from '../../../../../../../shared/src/util/errors'
import { propertyIsDefined } from '../../../../../../../shared/src/util/types'
import { queryGraphQL } from '../../../../../backend/graphql'
import { getCodeActions, getDiagnosticInfos } from '../../../../threadsOLD/detail/backend'
import { computeDiff } from '../../../../threadsOLD/detail/changes/computeDiff'

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
    }
`

const LOADING: 'loading' = 'loading'

type Result = typeof LOADING | GQL.ICampaignPreview | ErrorLike

/**
 * A React hook that observes a campaign preview queried from the GraphQL API.
 */
export const useCampaignPreview = (
    { extensionsController }: ExtensionsControllerProps,
    campaignInput: GQL.ICreateCampaignInput
): [Result, boolean] => {
    const [isLoading, setIsLoading] = useState(true)
    const [result, setResult] = useState<Result>(LOADING)
    useEffect(() => {
        const subscription = getDiagnosticInfos(extensionsController, 'eslint')
            .pipe(
                switchMap(diagnostics =>
                    combineLatest(
                        diagnostics.map(d =>
                            getCodeActions({
                                diagnostic: d,
                                extensionsController,
                            }).pipe(map(actions => ({ diagnostic: d, actions })))
                        )
                    )
                ),
                switchMap(diagnosticsAndActions => {
                    const fileDiffs = computeDiff(
                        extensionsController,
                        diagnosticsAndActions
                            .map(d => ({
                                actionEditCommand: d.actions[0].computeEdit,
                                diagnostic: fromDiagnostic(d.diagnostic),
                            }))
                            .filter(propertyIsDefined('actionEditCommand'))
                    )
                    return from(fileDiffs).pipe(
                        switchMap(fileDiffs =>
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
                                        campaign: campaignInput,
                                        rawDiagnostics: diagnosticsAndActions
                                            .map(({ diagnostic }) => diagnostic)
                                            .map(d =>
                                                // tslint:disable-next-line: no-object-literal-type-assertion
                                                JSON.stringify({
                                                    __typename: 'Diagnostic',
                                                    type: d.type,
                                                    data: d,
                                                } as GQL.IDiagnostic)
                                            ),
                                        rawFileDiffs: fileDiffs.map(({ patchWithFullURIs }) => patchWithFullURIs),
                                    },
                                } as GQL.ICampaignPreviewOnQueryArguments
                            ).pipe(
                                map(dataOrThrowErrors),
                                map(data => data.campaignPreview)
                            )
                        )
                    )
                }),
                startWith(LOADING)
            )
            .subscribe(
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
    }, [extensionsController, campaignInput])
    return [result, isLoading]
}
