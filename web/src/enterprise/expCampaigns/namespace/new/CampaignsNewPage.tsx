import { NotificationType } from '@sourcegraph/extension-api-classes'
import H from 'history'
import React, { useCallback, useEffect, useState } from 'react'
import { Redirect, RouteComponentProps } from 'react-router'
import { map } from 'rxjs/operators'
import { USE_CAMPAIGN_RULES } from '../..'
import { ExtensionsControllerProps } from '../../../../../../shared/src/extensions/controller'
import { dataOrThrowErrors, gql } from '../../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { PlatformContextProps } from '../../../../../../shared/src/platform/context'
import { mutateGraphQL } from '../../../../backend/graphql'
import { PageTitle } from '../../../../components/PageTitle'
import { ThemeProps } from '../../../../theme'
import { useLocalStorage } from '../../../../util/useLocalStorage'
import { getCompleteCampaignExtensionData } from '../../extensionData'
import { CampaignFormData } from '../../form/CampaignForm'
import { CampaignPreview } from '../../preview/CampaignPreview'
import { NamespaceCampaignsAreaContext } from '../NamespaceCampaignsArea'
import { NewCampaignForm } from './NewCampaignForm'
import { parseJSON } from '../../../../settings/configuration'
import { Workflow } from '../../../../schema/workflow.schema'
import { RULE_TEMPLATES } from '../../form/templates'
import { Markdown } from '../../../../../../shared/src/components/Markdown'
import { renderMarkdown } from '../../../../../../shared/src/util/markdown'

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
        RouteComponentProps<{ template: string }>,
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
export const CampaignsNewPage: React.FunctionComponent<Props> = ({ namespace, setBreadcrumbItem, match, ...props }) => {
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

    const template = RULE_TEMPLATES.find(t => t.id === match.params.template)

    // Persist the user's create-or-draft choice.
    const [defaultDraft, setDefaultDraft] = useLocalStorage('CampaignsNewPage.draft', true)

    const [value, setValue] = useLocalStorage<CampaignFormData>('CampaignsNewPage.value', {
        name: new URLSearchParams(props.location.search).get('name') || '',
        namespace: namespace.id,
        workflowAsJSONCString: template ? JSON.stringify(template.defaultWorkflow, null, 2) : '{}',
        draft: USE_CAMPAIGN_RULES ? defaultDraft : false,
    })
    const onChange = useCallback(
        (newValue: Partial<CampaignFormData>) => {
            if (newValue.draft !== undefined && newValue.draft !== defaultDraft) {
                setDefaultDraft(newValue.draft)
            }
            setValue(prevValue => ({ ...prevValue, ...newValue }))
        },
        [defaultDraft, setDefaultDraft, setValue]
    )

    const [creationOrError, setCreationOrError] = useState<null | typeof LOADING | Pick<GQL.IExpCampaign, 'url'>>(null)
    const onSubmit = useCallback(async () => {
        setCreationOrError(LOADING)
        try {
            const effectiveName = value.name || value.nameSuggestion || ''
            const extensionData = await getCompleteCampaignExtensionData(
                props.extensionsController,
                parseJSON(value.workflowAsJSONCString) as Workflow,
                { ...value, name: effectiveName }
            )
            setCreationOrError(
                await createCampaign({
                    ...value,
                    name: effectiveName,
                    namespace: namespace.id,
                    extensionData,
                })
            )
        } catch (err) {
            setCreationOrError(null)
            props.extensionsController.services.notifications.showMessages.next({
                message: `Error creating campaign: ${err.message}`,
                type: NotificationType.Error,
            })
        }
    }, [namespace.id, props.extensionsController, value])

    const isValid = value.name !== '' || value.nameSuggestion !== undefined

    const TemplateIcon = template ? template.icon : undefined

    return (
        <>
            {creationOrError !== null && creationOrError !== LOADING && <Redirect to={creationOrError.url} />}
            <PageTitle title="New campaign" />
            <div>
                {template ? (
                    <>
                        <h2 className="d-flex align-items-center">
                            {TemplateIcon && <TemplateIcon className="icon-inline mr-1" />} New campaign:{' '}
                            {template.title}
                        </h2>
                        {template.detail && (
                            <Markdown dangerousInnerHTML={renderMarkdown(template.detail)} className="text-muted" />
                        )}
                        <NewCampaignForm
                            {...props}
                            value={value}
                            isValid={isValid}
                            onChange={onChange}
                            onSubmit={onSubmit}
                            isLoading={creationOrError === LOADING}
                            workflowJSONSchema={template.workflowJSONSchema}
                            template={template}
                            className="flex-1"
                        />
                        {USE_CAMPAIGN_RULES && value && isValid && (
                            <>
                                <hr className="my-5" />
                                <CampaignPreview {...props} data={value} />
                            </>
                        )}
                    </>
                ) : (
                    <div className="alert alert-danger">
                        Template <code>{match.params.template}</code> not found.
                    </div>
                )}
            </div>
        </>
    )
}
