import H from 'history'
import React, { useCallback, useEffect, useMemo, useState } from 'react'
import { Redirect, RouteComponentProps } from 'react-router'
import { map } from 'rxjs/operators'
import { dataOrThrowErrors, gql } from '../../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { asError, ErrorLike, isErrorLike } from '../../../../../../shared/src/util/errors'
import { mutateGraphQL } from '../../../../backend/graphql'
import { PageTitle } from '../../../../components/PageTitle'
import { NamespaceCampaignsAreaContext } from '../NamespaceCampaignsArea'
import { CampaignForm, CampaignFormData } from './CampaignForm'
import { CampaignPreview } from './preview/CampaignPreview'
import { EMPTY_CAMPAIGN_TEMPLATE_ID } from './templates'

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

interface Props
    extends Pick<NamespaceCampaignsAreaContext, 'namespace' | 'setBreadcrumbItem'>,
        RouteComponentProps<{}> {
    location: H.Location
}

const LOADING = 'loading' as const

/**
 * Shows a form to create a new campaign.
 */
export const CampaignsNewPage: React.FunctionComponent<Props> = ({ namespace, setBreadcrumbItem, location, match }) => {
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

    const templateID = new URLSearchParams(location.search).get('template')
    const preview = !!templateID && templateID !== EMPTY_CAMPAIGN_TEMPLATE_ID

    const [data, setData] = useState<CampaignFormData>()
    const onChange = useCallback(data => setData(data), [])

    const [creationOrError, setCreationOrError] = useState<
        null | typeof LOADING | Pick<GQL.ICampaign, 'url'> | ErrorLike
    >(null)
    const onSubmit = useCallback(
        async (data: CampaignFormData) => {
            setCreationOrError(LOADING)
            try {
                setCreationOrError(await createCampaign({ ...data, namespace: namespace.id, preview }))
            } catch (err) {
                setCreationOrError(asError(err))
                alert(err.message) // TODO!(sqs)
            }
        },
        [namespace.id, preview]
    )

    const initialValue = useMemo<CampaignFormData>(() => ({ name: '', namespace: namespace.id }), [namespace.id])

    return (
        <>
            {creationOrError !== null && creationOrError !== LOADING && !isErrorLike(creationOrError) && (
                <Redirect to={creationOrError.url} />
            )}
            <PageTitle title="New campaign" />
            <div className="d-flex flex-gap-3 flex-wrap">
                <div style={{ flex: '1 0 50%', minWidth: '32rem' }}>
                    <CampaignForm
                        templateID={templateID}
                        initialValue={initialValue}
                        onChange={onChange}
                        onSubmit={onSubmit}
                        buttonText={preview ? 'Preview campaign' : 'Create campaign'}
                        isLoading={creationOrError === LOADING}
                        match={match}
                        location={location}
                        className="flex-1"
                    />
                </div>
                {/* TODO!(sqs): be smart about when to show the preview pane */}
                {templateID && templateID !== EMPTY_CAMPAIGN_TEMPLATE_ID && (
                    <div style={{ flex: '2 0 24rem' }}>{data && <CampaignPreview data={data} />}</div>
                )}
            </div>
            {isErrorLike(creationOrError) && <div className="alert alert-danger mt-3">{creationOrError.message}</div>}
        </>
    )
}
