import NpmIcon from 'mdi-react/NpmIcon'
import React, { useCallback, useEffect } from 'react'
import { CampaignTemplate, CampaignTemplateComponentContext } from '.'
import { PackageJsonDependencyCampaignContext } from '../../../../../../extensions/enterprise/sandbox/src/packageJsonDependency'
import { pluralize } from '../../../../../../shared/src/util/strings'
import { ParsedDiagnosticQuery, parseDiagnosticQuery } from '../../../diagnostics/diagnosticQuery'
import { RuleDefinition } from '../../../rules/types'
import { CampaignFormFiltersFormControl } from '../CampaignFormFiltersFormControl'

const TEMPLATE_ID = 'packageJsonDependency'

interface Props extends CampaignTemplateComponentContext {}

const ALL_VERSION_RANGE = '*'

const SAMPLE_PACKAGE_NAMES: { packageName: string; count: number }[] = [
    { packageName: 'typescript', count: 351 },
    { packageName: 'react', count: 91 },
    { packageName: 'lodash', count: 126 },
    { packageName: 'mdi-react', count: 53 },
    { packageName: 'glob', count: 29 },
    { packageName: '@sourcegraph/codeintellify', count: 15 },
]

const PackageJsonDependencyCampaignTemplateForm: React.FunctionComponent<Props> = ({
    value,
    onChange,
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
            const campaignName = `${newContext.ban ? 'Ban' : 'Deprecate'} ${newContext.packageName || '<package>'}${
                newContext.versionRange && newContext.versionRange !== ALL_VERSION_RANGE
                    ? `@${newContext.versionRange}`
                    : ''
            } (npm)`
            // TODO!(sqs): figure out how to make showWarnings also show warnings - is it a separate rule?
            onChange({
                isValid: !!newContext.packageName,
                name: campaignName,
                template: {
                    template: TEMPLATE_ID,
                    context: newContext,
                },
                rules: [
                    {
                        name: 'Remove dependency from package.json',
                        // eslint-disable-next-line @typescript-eslint/no-object-literal-type-assertion
                        definition: JSON.stringify({
                            type: 'DiagnosticRule',
                            query: diagnosticQuery('type:packageJsonDependency'),
                            context: newContext,
                            action: 'packageJsonDependency.remove',
                        } as RuleDefinition),
                    },
                ],
            })
        },
        [context, onChange]
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
            const versionRange = params.get('versionRange')
            if (versionRange !== null) {
                update.versionRange = versionRange
            }

            updateContext({
                versionRange: ALL_VERSION_RANGE,
                createChangesets: true,
                showWarnings: true,
                ban: false,
                ...update,
            })
        }
    }, [context, locationSearch, updateContext])

    const onPackageNameChange = useCallback<React.ChangeEventHandler<HTMLInputElement>>(
        e => updateContext({ packageName: e.currentTarget.value }),
        [updateContext]
    )

    const onVersionRangeChange = useCallback<React.ChangeEventHandler<HTMLInputElement>>(
        e => updateContext({ versionRange: e.currentTarget.value }),
        [updateContext]
    )

    const onCreateChangesetsChange = useCallback<React.ChangeEventHandler<HTMLInputElement>>(
        e => updateContext({ createChangesets: e.currentTarget.checked }),
        [updateContext]
    )

    const onShowWarningsChange = useCallback<React.ChangeEventHandler<HTMLInputElement>>(
        e => updateContext({ showWarnings: e.currentTarget.checked }),
        [updateContext]
    )

    const onBanChange = useCallback<React.ChangeEventHandler<HTMLInputElement>>(
        e => updateContext({ ban: e.currentTarget.checked }),
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
                    list="campaign-template-form__packageName-datalist"
                />
                <datalist id="campaign-template-form__packageName-datalist">
                    {SAMPLE_PACKAGE_NAMES.map(({ packageName, count }) => (
                        <option key={packageName} value={packageName}>
                            {count} {pluralize('dependent', count)}
                        </option>
                    ))}
                </datalist>
            </div>
            <div className="form-group">
                <label htmlFor="campaign-template-form__versionRange">Version range (to deprecate)</label>
                <input
                    type="text"
                    id="campaign-template-form__versionRange"
                    className="form-control"
                    placeholder="semver range (e.g., <1.2.3)"
                    value={context.versionRange || ''}
                    onChange={onVersionRangeChange}
                    disabled={disabled}
                />
                <p className="form-help text-muted small mb-0">
                    <a href="https://docs.npmjs.com/misc/semver#ranges" target="_blank" rel="noopener noreferrer">
                        How to specify version ranges
                    </a>{' '}
                    (<code>{ALL_VERSION_RANGE}</code> matches all versions)
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
                    <li className="form-check">
                        <input
                            type="checkbox"
                            id="campaign-template-form__showWarnings"
                            className="form-check-input"
                            checked={context.showWarnings}
                            onChange={onShowWarningsChange}
                            disabled={disabled}
                        />
                        <label className="form-check-label" htmlFor="campaign-template-form__showWarnings">
                            Show diagnostics on all active branches
                        </label>
                    </li>
                    <li className="form-check">
                        <input
                            type="checkbox"
                            id="campaign-template-form__ban"
                            className="form-check-input"
                            checked={context.ban}
                            onChange={onBanChange}
                            disabled={disabled}
                        />
                        <label className="form-check-label" htmlFor="campaign-template-form__ban">
                            Ban (immediately fail all builds, including default branches, with this dependency version)
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

export const PackageJsonDependencyCampaignTemplate: CampaignTemplate = {
    id: TEMPLATE_ID,
    title: 'package.json dependency deprecation/ban',
    detail:
        'Deprecate or ban an npm/yarn dependency in package.json manifests, opening issues/changesets for all affected code owners.',
    icon: NpmIcon,
    renderForm: PackageJsonDependencyCampaignTemplateForm,
}
