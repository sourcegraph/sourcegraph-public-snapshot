import SecurityLockIcon from 'mdi-react/SecurityLockIcon'
import React, { useCallback, useEffect } from 'react'
import { RuleTemplate, RuleTemplateComponentContext } from '.'
import { NPMCredentialsCampaignContext } from '../../../../../../extensions/enterprise/sandbox/src/npmCredentials/providers'
import { ParsedDiagnosticQuery, parseDiagnosticQuery } from '../../../diagnostics/diagnosticQuery'
import { RuleDefinition } from '../../../rules/types'
import { CampaignFormFiltersFormControl } from '../CampaignFormFiltersFormControl'

const TEMPLATE_ID = 'npmCredentials'

interface Props extends RuleTemplateComponentContext {}

const NPMCredentialsCampaignTemplateForm: React.FunctionComponent<Props> = ({
    value,
    onChange,
    onCampaignChange,
    disabled,
    location,
}) => {
    const context: NPMCredentialsCampaignContext | undefined = value.template ? value.template.context : undefined

    const updateContext = useCallback(
        (update: Partial<NPMCredentialsCampaignContext>): void => {
            const newContext = { ...context, ...update }
            const diagnosticQuery = (query: string): ParsedDiagnosticQuery =>
                parseDiagnosticQuery(`${newContext.filters || ''}${newContext.filters ? ' ' : ''}${query}`)
            const campaignName = 'Find npm credentials'
            onCampaignChange({
                isValid: true,
                name: campaignName,
            })
            onChange({
                name: 'Remove npm credentials',
                template: {
                    template: TEMPLATE_ID,
                    context: newContext,
                },
                // eslint-disable-next-line @typescript-eslint/no-object-literal-type-assertion
                definition: JSON.stringify({
                    type: 'DiagnosticRule',
                    query: diagnosticQuery('type:npmCredentials'),
                    context: newContext,
                    action: 'npmCredentials.fix',
                } as RuleDefinition),
            })
        },
        [context, onCampaignChange, onChange]
    )

    // Set defaults.
    const locationSearch = location ? location.search : ''
    useEffect(() => {
        if (context === undefined) {
            const update: Partial<NPMCredentialsCampaignContext> = {}

            const params = new URLSearchParams(locationSearch)
            const filters = params.get('filters')
            if (filters !== null) {
                update.filters = filters
            }

            updateContext({
                ...update,
            })
        }
    }, [context, locationSearch, updateContext])

    const onFiltersChange = useCallback((value: string) => updateContext({ filters: value }), [updateContext])

    if (context === undefined) {
        return null
    }

    return (
        <>
            <CampaignFormFiltersFormControl
                value={context.filters || ''}
                onChange={onFiltersChange}
                disabled={disabled}
            />
        </>
    )
}

export const NPMCredentialsRuleTemplate: RuleTemplate = {
    id: TEMPLATE_ID,
    title: 'npm credentials',
    detail: 'Find npm credentials, revoke them on Artifactory, and open a pull request to remove them.',
    icon: SecurityLockIcon,
    renderForm: NPMCredentialsCampaignTemplateForm,
}
