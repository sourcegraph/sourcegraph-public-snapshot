import NpmIcon from 'mdi-react/NpmIcon'
import React, { useCallback, useEffect } from 'react'
import { RuleTemplate, RuleTemplateComponentContext } from '.'
import { PackageJsonDependencyCampaignContext } from '../../../../../../extensions/enterprise/sandbox/src/packageJsonDependency/packageJsonDependency'
import { ParsedDiagnosticQuery, parseDiagnosticQuery } from '../../../diagnostics/diagnosticQuery'
import { RuleDefinition } from '../../../rules/types'
import { CampaignFormFiltersFormControl } from '../CampaignFormFiltersFormControl'

const TEMPLATE_ID = 'packageJsonDependency'

interface Props extends RuleTemplateComponentContext {}

const PackageJsonDependencyCampaignTemplateForm: React.FunctionComponent<Props> = ({
    value,
    onChange,
    onCampaignChange,
    disabled,
    location,
}) => {
    const context: PackageJsonDependencyCampaignContext | undefined = value.template
        ? value.template.context
        : undefined

    const updateContext = useCallback(
        (update: Partial<PackageJsonDependencyCampaignContext>): void => {
            const newContext = { ...context, ...update }
            const diagnosticQuery = (query: string): ParsedDiagnosticQuery =>
                parseDiagnosticQuery(`${newContext.filters || ''}${newContext.filters ? ' ' : ''}${query}`)
            const campaignName = `Upgrade npm dependency ${newContext.packageName ||
                '<package>'} to ${newContext.upgradeToVersion || '<version>'}`
            onCampaignChange({
                isValid: !!newContext.packageName,
                name: campaignName,
            })
            onChange({
                name: 'Remove dependency from package.json',
                template: {
                    template: TEMPLATE_ID,
                    context: newContext,
                },
                // eslint-disable-next-line @typescript-eslint/no-object-literal-type-assertion
                definition: JSON.stringify({
                    type: 'DiagnosticRule',
                    query: diagnosticQuery('type:packageJsonDependency'),
                    context: newContext,
                    action: 'packageJsonDependency.upgrade',
                } as RuleDefinition),
            })
        },
        [context, onCampaignChange, onChange]
    )

    // Set defaults.
    const locationSearch = location ? location.search : ''
    useEffect(() => {
        if (context === undefined) {
            const update: Partial<PackageJsonDependencyCampaignContext> = {}

            const params = new URLSearchParams(locationSearch)
            const packageName = params.get('packageName')
            if (packageName !== null) {
                update.packageName = packageName
            }
            const upgradeToVersion = params.get('upgradeToVersion')
            if (upgradeToVersion !== null) {
                update.upgradeToVersion = upgradeToVersion
            }
            const filters = params.get('filters')
            if (filters !== null) {
                update.filters = filters
            }

            updateContext({
                createChangesets: true,
                ...update,
            })
        }
    }, [context, locationSearch, updateContext])

    const onPackageNameChange = useCallback<React.ChangeEventHandler<HTMLInputElement>>(
        e => updateContext({ packageName: e.currentTarget.value }),
        [updateContext]
    )

    const onUpgradeToVersionChange = useCallback<React.ChangeEventHandler<HTMLInputElement>>(
        e => updateContext({ upgradeToVersion: e.currentTarget.value }),
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
                <label htmlFor="campaign-template-form__packageName">Package name</label>
                <input
                    type="text"
                    id="campaign-template-form__packageName"
                    className="form-control"
                    required={true}
                    minLength={1}
                    placeholder="npm package name (e.g., lodash)"
                    value={context.packageName || ''}
                    onChange={onPackageNameChange}
                    autoFocus={true}
                    disabled={disabled}
                />
            </div>
            <div className="form-group">
                <label htmlFor="campaign-template-form__upgradeToVersion">Upgrade to version</label>
                <input
                    type="text"
                    id="campaign-template-form__upgradeToVersion"
                    className="form-control"
                    placeholder="semver range (e.g., <1.2.3)"
                    value={context.upgradeToVersion || ''}
                    onChange={onUpgradeToVersionChange}
                    disabled={disabled}
                />
                <p className="form-help text-muted small mb-0 mt-1">
                    <a href="https://semver.npmjs.com/" target="_blank" rel="noopener noreferrer">
                        Version range calculator
                    </a>{' '}
                    &bull; Examples: <code className="border-bottom mr-3">&gt;=1.10.0</code>
                    <code className="border-bottom mr-3">~0.2.2 || ^0.3.2</code>
                </p>
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
                            Create changesets with dependency removed from package.json
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

export const PackageJsonDependencyRuleTemplate: RuleTemplate = {
    id: TEMPLATE_ID,
    title: 'package.json dependency upgrade',
    detail:
        'Upgrade an npm/yarn dependency in package.json files, opening issues/changesets for all affected code owners.',
    icon: NpmIcon,
    renderForm: PackageJsonDependencyCampaignTemplateForm,
}
