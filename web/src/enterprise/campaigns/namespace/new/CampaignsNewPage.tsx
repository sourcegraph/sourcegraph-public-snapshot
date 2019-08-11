import { NotificationType } from '@sourcegraph/extension-api-classes'
import H from 'history'
import React, { useCallback, useEffect, useMemo, useState } from 'react'
import { Redirect, RouteComponentProps } from 'react-router'
import { first, map } from 'rxjs/operators'
import { ExtensionsControllerProps } from '../../../../../../shared/src/extensions/controller'
import { dataOrThrowErrors, gql } from '../../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { PlatformContextProps } from '../../../../../../shared/src/platform/context'
import { mutateGraphQL } from '../../../../backend/graphql'
import { PageTitle } from '../../../../components/PageTitle'
import { ThemeProps } from '../../../../theme'
import { getCampaignExtensionData } from '../../extensionData'
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
        RouteComponentProps<{}>,
        ExtensionsControllerProps,
        PlatformContextProps,
        ThemeProps {
    location: H.Location
    history: H.History
}

const LOADING = 'loading' as const

/**
 * Shows a form to create a new campaign.
 */
export const CampaignsNewPage: React.FunctionComponent<Props> = ({
    namespace,
    setBreadcrumbItem,
    location,
    match,
    ...props
}) => {
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

    const [data, setData] = useState<CampaignFormData>()
    const onChange = useCallback(data => setData(data), [])

    const [creationOrError, setCreationOrError] = useState<null | typeof LOADING | Pick<GQL.ICampaign, 'url'>>(null)
    const onSubmit = useCallback(
        async (data: CampaignFormData) => {
            setCreationOrError(LOADING)
            try {
                const extensionData = await getCampaignExtensionData(props.extensionsController, data)
                    .pipe(first())
                    .toPromise()
                setCreationOrError(await createCampaign({ ...data, namespace: namespace.id, extensionData }))
            } catch (err) {
                setCreationOrError(null)
                props.extensionsController.services.notifications.showMessages.next({
                    message: `Error creating campaign: ${err.message}`,
                    type: NotificationType.Error,
                })
            }
        },
        [namespace.id, props.extensionsController]
    )

    const initialValue = useMemo<CampaignFormData>(() => ({ name: '', namespace: namespace.id, isValid: true }), [
        namespace.id,
    ])

    return (
        <>
            {creationOrError !== null && creationOrError !== LOADING && <Redirect to={creationOrError.url} />}
            <PageTitle title="New campaign" />
            <div>
                <CampaignForm
                    templateID={templateID}
                    initialValue={initialValue}
                    onChange={onChange}
                    onSubmit={onSubmit}
                    buttonText="Create campaign"
                    isLoading={creationOrError === LOADING}
                    match={match}
                    location={location}
                    className="flex-1"
                />
                {/* TODO!(sqs): be smart about when to show the preview pane */}
                {templateID && templateID !== EMPTY_CAMPAIGN_TEMPLATE_ID && data && data.isValid && (
                    <>
                        <hr className="my-5" />
                        <CampaignPreview {...props} data={data} location={location} />
                    </>
                )}
            </div>
        </>
    )
}
