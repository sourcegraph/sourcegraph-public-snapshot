import { useEffect, useState } from 'react'
import { map, startWith, switchMap } from 'rxjs/operators'
import { ExtensionsControllerProps } from '../../../../../../../shared/src/extensions/controller'
import { dataOrThrowErrors, gql } from '../../../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../../../shared/src/graphql/schema'
import { asError, ErrorLike, isErrorLike } from '../../../../../../../shared/src/util/errors'
import { queryGraphQL } from '../../../../../backend/graphql'

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
        const subscription = extensionsController.services.diagnostics
            .observeDiagnostics({}, 'eslint')
            .pipe(
                map(diagnostics =>
                    diagnostics.map(d =>
                        // tslint:disable-next-line: no-object-literal-type-assertion
                        JSON.stringify({
                            __typename: 'Diagnostic',
                            type: d.type,
                            data: d,
                        } as GQL.IDiagnostic)
                    )
                ),
                switchMap(rawDiagnostics =>
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
                            input: { campaign: campaignInput, rawDiagnostics, rawFileDiffs: ['TODO!(sqs)'] },
                        } as GQL.ICampaignPreviewOnQueryArguments
                    )
                ),
                map(dataOrThrowErrors),
                map(data => data.campaignPreview),
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
