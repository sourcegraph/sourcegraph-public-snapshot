import React, { useCallback, useEffect, useState } from 'react'
import { Redirect } from 'react-router'
import { map } from 'rxjs/operators'
import { dataOrThrowErrors, gql } from '../../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { asError, ErrorLike, isErrorLike } from '../../../../../../shared/src/util/errors'
import { mutateGraphQL } from '../../../../backend/graphql'
import { ModalPage } from '../../../../components/ModalPage'
import { PageTitle } from '../../../../components/PageTitle'
import { CampaignForm, CampaignFormData } from '../../form/CampaignForm'
import { NamespaceCampaignsAreaContext } from '../NamespaceCampaignsArea'

export const createCampaign = (input: GQL.ICreateCampaignInput): Promise<GQL.ICampaign> =>
    mutateGraphQL(
        gql`
            mutation CreateCampaign($input: CreateCampaignInput!) {
                createCampaign(input: $input) {
                    id
                    url
                }
            }
        `,
        { input }
    )
        .pipe(
            map(dataOrThrowErrors),
            map(data => data.createCampaign)
        )
        .toPromise()

interface Props extends Pick<NamespaceCampaignsAreaContext, 'namespace' | 'setBreadcrumbItem'> {}

const LOADING = 'loading' as const

/**
 * Shows a form to create a new campaign.
 */
export const CampaignsNewPage: React.FunctionComponent<Props> = ({ namespace, setBreadcrumbItem }) => {
    useEffect(() => {
        if (setBreadcrumbItem) {
            setBreadcrumbItem({ text: 'New' })
        }
        return () => {
            if (setBreadcrumbItem) {
                setBreadcrumbItem(undefined)
            }
        }
    }, [setBreadcrumbItem])

    const [creationOrError, setCreationOrError] = useState<
        null | typeof LOADING | Pick<GQL.ICampaign, 'url'> | ErrorLike
    >(null)
    const onSubmit = useCallback(
        async (data: CampaignFormData) => {
            setCreationOrError(LOADING)
            try {
                setCreationOrError(await createCampaign({ ...data, namespace: namespace.id }))
            } catch (err) {
                setCreationOrError(asError(err))
                alert(err.message) // TODO!(sqs)
            }
        },
        [namespace.id]
    )

    return (
        <>
            {creationOrError !== null && creationOrError !== LOADING && !isErrorLike(creationOrError) && (
                <Redirect to={creationOrError.url} />
            )}
            <PageTitle title="New campaign" />
            <ModalPage>
                <h2>New campaign</h2>
                <CampaignForm
                    onSubmit={onSubmit}
                    buttonText="Create campaign"
                    isLoading={creationOrError === LOADING}
                />
                {isErrorLike(creationOrError) && (
                    <div className="alert alert-danger mt-3">{creationOrError.message}</div>
                )}
            </ModalPage>
        </>
    )
}
