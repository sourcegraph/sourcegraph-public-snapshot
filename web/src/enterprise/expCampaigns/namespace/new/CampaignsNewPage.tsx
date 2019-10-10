import { NotificationType } from '@sourcegraph/extension-api-classes'
import H from 'history'
import React, { useCallback, useEffect, useMemo, useState } from 'react'
import { Redirect, RouteComponentProps } from 'react-router'
import { first, map } from 'rxjs/operators'
import { USE_CAMPAIGN_RULES } from '../..'
import { ExtensionsControllerProps } from '../../../../../../shared/src/extensions/controller'
import { dataOrThrowErrors, gql } from '../../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { PlatformContextProps } from '../../../../../../shared/src/platform/context'
import { mutateGraphQL } from '../../../../backend/graphql'
import { PageTitle } from '../../../../components/PageTitle'
import { ThemeProps } from '../../../../theme'
import { useLocalStorage } from '../../../../util/useLocalStorage'
import { RuleDefinition } from '../../../rules/types'
import { getCampaignExtensionData } from '../../extensionData'
import { CampaignFormData } from '../../form/CampaignForm'
import { EMPTY_RULE_TEMPLATE_ID } from '../../form/templates'
import { CampaignPreview } from '../../preview/CampaignPreview'
import { NamespaceCampaignsAreaContext } from '../NamespaceCampaignsArea'
import { NewCampaignForm } from './NewCampaignForm'

export const createCampaign = (input: GQL.IExpCreateCampaignInput): Promise<GQL.IExpCampaign> =>
    mutateGraphQL(
        gql`
            mutation CreateCampaign($input: ExpCreateCampaignInput!) {
                expCreateCampaign(input: $input) {
                    id
                    url
                }
            }
        `,
        { input }
    )
        .pipe(
            map(dataOrThrowErrors),
            map(data => data.expCreateCampaign)
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

    const ruleTemplateID = new URLSearchParams(location.search).get('template')
    const defaultRules = useMemo<CampaignFormData['rules']>(
        () =>
            ruleTemplateID
                ? [{ name: '', template: { template: ruleTemplateID, context: undefined }, definition: '{}' }]
                : undefined,
        [ruleTemplateID]
    )

    // Persist the user's create-or-draft choice.
    const [defaultDraft, setDefaultDraft] = useLocalStorage('CampaignsNewPage.draft', true)

    const [value, setValue] = useState<CampaignFormData>({
        name: '',
        namespace: namespace.id,
        draft: USE_CAMPAIGN_RULES ? defaultDraft : false,
        isValid: true,
    })
    useEffect(() => setValue(prevValue => ({ ...prevValue, rules: defaultRules })), [defaultRules])
    const onChange = useCallback(
        (newValue: Partial<CampaignFormData>) => {
            if (newValue.draft !== undefined && newValue.draft !== defaultDraft) {
                setDefaultDraft(newValue.draft)
            }
            setValue(prevValue => ({ ...prevValue, ...newValue }))
        },
        [defaultDraft, setDefaultDraft]
    )

    const [creationOrError, setCreationOrError] = useState<null | typeof LOADING | Pick<GQL.IExpCampaign, 'url'>>(null)
    const onSubmit = useCallback(async () => {
        setCreationOrError(LOADING)
        try {
            const [extensionData] = await getCampaignExtensionData(
                props.extensionsController,
                value.rules ? value.rules.map(rule => JSON.parse(rule.definition) as RuleDefinition) : []
            )
                .pipe(first(xd => !xd[1].isLoading))
                .toPromise()
            setCreationOrError(await createCampaign({ ...value, namespace: namespace.id, extensionData }))
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
                    match={match}
                    location={location}
                    className="flex-1"
                />
                {USE_CAMPAIGN_RULES &&
                    ruleTemplateID &&
                    ruleTemplateID !== EMPTY_RULE_TEMPLATE_ID &&
                    value &&
                    value.isValid && (
                        <>
                            <hr className="my-5" />
                            <CampaignPreview {...props} data={value} location={location} />
                        </>
                    )}
            </div>
        </>
    )
}
