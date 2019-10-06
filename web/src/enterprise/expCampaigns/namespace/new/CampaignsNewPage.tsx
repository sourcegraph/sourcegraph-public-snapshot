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

    // Persist the user's create-or-draft choice.
    const [defaultDraft, setDefaultDraft] = useLocalStorage('CampaignsNewPage.draft', true)

    const [value, setValue] = useLocalStorage<CampaignFormData>('CampaignsNewPage.value', {
        name: new URLSearchParams(props.location.search).get('name') || '',
        namespace: namespace.id,
        workflowAsJSONCString: JSON.stringify(
            // eslint-disable-next-line @typescript-eslint/no-object-literal-type-assertion
            {
                // extensions: ['sourcegraph/automation-preview'],
                variables: {
                    packageName: 'react-router-dom',
                    matchVersion: '*',
                    action: { requireVersion: '^5.0.1' },
                    createChangesets: true,
                    headBranch: 'upgrade-react-router-dom',
                },
                run: [
                    {
                        diagnostics: { id: 'dependencyManagement.packageJsonDependency' },
                        codeActions: [{ command: 'dependencyManagement.packageJsonDependency.action' }],
                    },
                ],
                behaviors: {
                    edits: { command: 'changesets.byRepositoryAndBaseBranch' },
                },
            } as Workflow,
            null,
            2
        ),
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
            const extensionData = await getCompleteCampaignExtensionData(
                props.extensionsController,
                parseJSON(value.workflowAsJSONCString) as Workflow,
                value
            )
            setCreationOrError(await createCampaign({ ...value, namespace: namespace.id, extensionData }))
        } catch (err) {
            setCreationOrError(null)
            props.extensionsController.services.notifications.showMessages.next({
                message: `Error creating campaign: ${err.message}`,
                type: NotificationType.Error,
            })
        }
    }, [namespace.id, props.extensionsController, value])

    const isValid = value.name !== ''

    return (
        <>
            {creationOrError !== null && creationOrError !== LOADING && <Redirect to={creationOrError.url} />}
            <PageTitle title="New campaign" />
            <div>
                <NewCampaignForm
                    {...props}
                    value={value}
                    isValid={isValid}
                    onChange={onChange}
                    onSubmit={onSubmit}
                    isLoading={creationOrError === LOADING}
                    className="flex-1"
                />
                {USE_CAMPAIGN_RULES && value && isValid && (
                    <>
                        <hr className="my-5" />
                        <CampaignPreview {...props} data={value} />
                    </>
                )}
            </div>
        </>
    )
}
