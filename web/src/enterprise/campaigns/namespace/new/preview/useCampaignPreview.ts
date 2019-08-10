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
    }
`

const LOADING: 'loading' = 'loading'

type Result = typeof LOADING | GQL.ICampaignPreview | ErrorLike

/**
 * A React hook that observes a campaign preview queried from the GraphQL API.
 */
export const useCampaignPreview = (
    { extensionsController }: ExtensionsControllerProps,
    input: GQL.ICreateCampaignInput
): [Result, boolean] => {
    const [isLoading, setIsLoading] = useState(true)
    const [result, setResult] = useState<Result>(LOADING)
    useEffect(() => {
        const subscription = queryGraphQL(
            gql`
                query CampaignPreview($input: CreateCampaignInput!) {
                    campaignPreview(input: $input) {
                        ...CampaignPreviewFragment
                    }
                }
                ${CampaignPreviewFragment}
            `,
            { input }
        )
            .pipe(
                map(dataOrThrowErrors),
                map(data => data.campaignPreview),
                switchMap(campaignPreview =>
                    extensionsController.services.diagnostics.observeDiagnostics({}, 'eslint').pipe(
                        map(diagnostics => ({
                            ...campaignPreview,
                            // tslint:disable-next-line: no-object-literal-type-assertion
                            diagnostics: {
                                __typename: 'DiagnosticConnection',
                                nodes: diagnostics.map(
                                    d =>
                                        // tslint:disable-next-line: no-object-literal-type-assertion
                                        ({
                                            __typename: 'Diagnostic',
                                            type: d.type,
                                            data: JSON.stringify(d),
                                        } as GQL.IDiagnostic)
                                ),
                                totalCount: diagnostics.length,
                                pageInfo: { hasNextPage: false },
                            } as GQL.IDiagnosticConnection,
                        }))
                    )
                )
            )
            .pipe(startWith(LOADING))
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
                err => setResult(asError(err))
            )
        return () => subscription.unsubscribe()
    }, [extensionsController, input])
    return [result, isLoading]
}
