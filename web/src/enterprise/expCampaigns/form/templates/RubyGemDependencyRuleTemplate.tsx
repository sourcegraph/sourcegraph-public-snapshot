import RubyIcon from 'mdi-react/RubyIcon'
import React, { useCallback, useEffect } from 'react'
import { RuleTemplate, RuleTemplateComponentContext } from '.'
import { RubyGemDependencyCampaignContext } from '../../../../../../extensions/enterprise/sandbox/src/rubyGemDependency'
import { pluralize } from '../../../../../../shared/src/util/strings'
import { ParsedDiagnosticQuery, parseDiagnosticQuery } from '../../../diagnostics/diagnosticQuery'
import { RuleDefinition } from '../../../rules/types'
import { CampaignFormFiltersFormControl } from '../CampaignFormFiltersFormControl'
import InformationOutlineIcon from 'mdi-react/InformationOutlineIcon'

const TEMPLATE_ID = 'rubyGemDependency'

interface Props extends RuleTemplateComponentContext {}

const ALL_VERSION_RANGE = '*'

const SAMPLE_PACKAGE_NAMES: { gemName: string; count: number }[] = [
    { gemName: 'rails', count: 351 },
    { gemName: 'rake', count: 91 },
    { gemName: 'guard', count: 126 },
    { gemName: 'guard-rspec', count: 53 },
    { gemName: 'omniauth', count: 29 },
    { gemName: 'omniauth-openid', count: 15 },
]

const RubyGemDependencyRuleTemplateForm: React.FunctionComponent<Props> = ({
    value,
    onChange,
    onCampaignChange,
    disabled,
    location,
}) => {
    const context: RubyGemDependencyCampaignContext | undefined = value.template ? value.template.context : undefined

    const updateContext = useCallback(
        (update: Partial<RubyGemDependencyCampaignContext>): void => {
            const newContext = { ...context, ...update }
            const diagnosticQuery = (query: string): ParsedDiagnosticQuery =>
                parseDiagnosticQuery(`${newContext.filters || ''}${newContext.filters ? ' ' : ''}${query}`)
            const campaignName = `Ban Ruby gem ${newContext.gemName || '<name>'} (Ruby)`
            onCampaignChange({
                isValid: !!newContext.gemName,
                name: campaignName,
            })
            onChange({
                name: 'Ban Ruby gem dependency',
                template: {
                    template: TEMPLATE_ID,
                    context: newContext,
                },
                // eslint-disable-next-line @typescript-eslint/no-object-literal-type-assertion
                definition: JSON.stringify({
                    type: 'DiagnosticRule',
                    query: diagnosticQuery('type:rubyGemDependency'),
                    context: newContext,
                    action: 'rubyGemDependency.remove',
                } as RuleDefinition),
            })
        },
        [context, onCampaignChange, onChange]
    )

    // Set defaults.
    const locationSearch = location ? location.search : ''
    useEffect(() => {
        if (context === undefined) {
            const update: Partial<RubyGemDependencyCampaignContext> = {}

            const params = new URLSearchParams(locationSearch)
            const gemName = params.get('gemName')
            if (gemName !== null) {
                update.gemName = gemName
            }

            updateContext({
                createChangesets: true,
                ...update,
            })
        }
    }, [context, locationSearch, updateContext])

    const onGemNameChange = useCallback<React.ChangeEventHandler<HTMLInputElement>>(
        e => updateContext({ gemName: e.currentTarget.value }),
        [updateContext]
    )

    const onCreateChangesetsChange = useCallback<React.ChangeEventHandler<HTMLInputElement>>(
        e => updateContext({ createChangesets: e.currentTarget.checked }),
        [updateContext]
    )

    const onFiltersChange = useCallback((value: string) => updateContext({ filters: value }), [updateContext])

    if (context === undefined) {
        return null
    }

    return (
        <>
            <div className="form-group">
                <label htmlFor="campaign-template-form__gemName">Gem name</label>
                <input
                    type="text"
                    id="campaign-template-form__gemName"
                    className="form-control"
                    required={true}
                    minLength={1}
                    placeholder="Ruby gem name (e.g., rails)"
                    value={context.gemName || ''}
                    onChange={onGemNameChange}
                    autoFocus={true}
                    disabled={disabled}
                    list="campaign-template-form__gemName-datalist"
                />
                <datalist id="campaign-template-form__gemName-datalist">
                    {SAMPLE_PACKAGE_NAMES.map(({ gemName, count }) => (
                        <option key={gemName} value={gemName}>
                            {count} {pluralize('dependent', count)}
                        </option>
                    ))}
                </datalist>
            </div>
            <div className="form-group">
                <label>Options</label>
                <ul className="list-unstyled">
                    <li className="form-check">
                        <input
                            type="checkbox"
                            id="campaign-template-form__createChangesets"
                            className="form-check-input"
                            checked={context.createChangesets}
                            onChange={onCreateChangesetsChange}
                            disabled={disabled}
                        />
                        <label className="form-check-label" htmlFor="campaign-template-form__createChangesets">
                            Create changesets (instead of just creating issues, where possible) to remove dependency
                            from <code>Gemfile</code> and <code>Gemfile.lock</code> files and <code>require</code>{' '}
                            statements{' '}
                            <span data-tooltip="Manual followup is usually needed (e.g., to remove calls to the removed gem's functions).">
                                <InformationOutlineIcon className="icon-inline" />
                            </span>
                        </label>
                    </li>
                </ul>
            </div>
            <CampaignFormFiltersFormControl
                value={context.filters || ''}
                onChange={onFiltersChange}
                disabled={disabled}
            />
        </>
    )
}

export const RubyGemDependencyRuleTemplate: RuleTemplate = {
    id: TEMPLATE_ID,
    title: 'Ruby gem dependency upgrade/deprecation',
    detail:
        'Modifies Gemfile and Gemfile.lock files. Issues or changesets will be opened on all repositories using this dependency.',
    icon: RubyIcon,
    renderForm: RubyGemDependencyRuleTemplateForm,
}
