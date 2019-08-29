import { NotificationType } from '@sourcegraph/extension-api-classes'
import React, { useCallback, useEffect, useMemo, useState } from 'react'
import { Redirect, RouteComponentProps } from 'react-router'
import { map } from 'rxjs/operators'
import { ExtensionsControllerProps } from '../../../../../../shared/src/extensions/controller'
import { dataOrThrowErrors, gql } from '../../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { PlatformContextProps } from '../../../../../../shared/src/platform/context'
import { mutateGraphQL } from '../../../../backend/graphql'
import { PageTitle } from '../../../../components/PageTitle'
import { ThemeProps } from '../../../../theme'
import { NamespaceCampaignsAreaContext } from '../NamespaceCampaignsArea'
import { CampaignFormData } from '../../form/CampaignForm'
import { NewCampaignForm } from './NewCampaignForm'

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
        ThemeProps {}

const LOADING = 'loading' as const

/**
 * Shows a form to create a new campaign.
 */
export const CampaignsNewPage: React.FunctionComponent<Props> = ({ namespace, setBreadcrumbItem, ...props }) => {
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

    const initialValue = useMemo<CampaignFormData>(() => ({ name: '', namespace: namespace.id, isValid: true }), [
        namespace.id,
    ])
    const [value, setValue] = useState<CampaignFormData>(initialValue)
    const onChange = useCallback((newValue: Partial<CampaignFormData>) => {
        setValue(prevValue => ({ ...prevValue, ...newValue }))
    }, [])

    const [creationOrError, setCreationOrError] = useState<null | typeof LOADING | Pick<GQL.ICampaign, 'url'>>(null)
    const onSubmit = useCallback(async () => {
        setCreationOrError(LOADING)
        try {
            setCreationOrError(await createCampaign({ ...value, namespace: namespace.id }))
        } catch (err) {
            setCreationOrError(null)
            props.extensionsController.services.notifications.showMessages.next({
                message: `Error creating campaign: ${err.message}`,
                type: NotificationType.Error,
            })
        }
    }, [namespace.id, props.extensionsController, value])

    return (
        <>
            {creationOrError !== null && creationOrError !== LOADING && <Redirect to={creationOrError.url} />}
            <PageTitle title="New campaign" />
            <div>
                <NewCampaignForm
                    value={value}
                    onChange={onChange}
                    onSubmit={onSubmit}
                    isLoading={creationOrError === LOADING}
                    className="flex-1"
                />
            </div>
        </>
    )
}
