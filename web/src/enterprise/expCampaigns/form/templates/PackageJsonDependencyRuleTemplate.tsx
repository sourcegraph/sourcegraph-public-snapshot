import NpmIcon from 'mdi-react/NpmIcon'
import React, { useCallback, useEffect, useState } from 'react'
import { RuleTemplate, RuleTemplateComponentContext } from '.'
import { PackageJsonDependencyCampaignContext } from '../../../../../../extensions/enterprise/sandbox/src/packageJsonDependency'
import { ParsedDiagnosticQuery, parseDiagnosticQuery } from '../../../diagnostics/diagnosticQuery'
import { RuleDefinition } from '../../../rulesOLD/types'
import { CampaignFormFiltersFormControl } from '../CampaignFormFiltersFormControl'
import { JSONSchema7 } from 'json-schema'

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
            const diagnosticQuery = (query: string): ParsedDiagnosticQuery => parseDiagnosticQuery(query)
            const packageNameAndVersion = `${newContext.packageName || '<package>'}${
                newContext.matchVersion ? `@${newContext.matchVersion}` : ''
            }`
            const campaignName =
                newContext.action === 'ban'
                    ? `Ban npm dependency ${packageNameAndVersion}`
                    : `Upgrade npm dependency ${packageNameAndVersion} to ${
                          newContext.action && newContext.action.requireVersion
                              ? newContext.action.requireVersion
                              : '<version>'
                      }`
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
                    action: 'dependencyManagement.packageJsonDependency.action',
                } as RuleDefinition),
            })
        },
        [context, onCampaignChange, onChange]
    )

    const [actionRequireVersion, setActionRequireVersion] = useState<string>('')

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

            const matchVersion = params.get('matchVersion')
            if (matchVersion !== null) {
                update.matchVersion = matchVersion
            }

            if (params.has('ban')) {
                update.action = 'ban'
            } else {
                const requireVersion = params.get('requireVersion')
                if (requireVersion) {
                    update.action = { requireVersion }
                    setActionRequireVersion(requireVersion)
                }
            }

            const createChangesets = params.get('createChangesets')
            if (createChangesets !== null) {
                update.createChangesets =
                    createChangesets === 'true' || createChangesets === '1' || createChangesets === 'yes'
            }

            const filters = params.get('filters')
            if (filters !== null) {
                update.filters = filters
            }

            updateContext({
                matchVersion: '*',
                ...update,
            })
        }
    }, [context, locationSearch, updateContext])

    const onPackageNameChange = useCallback<React.ChangeEventHandler<HTMLInputElement>>(
        e => updateContext({ packageName: e.currentTarget.value }),
        [updateContext]
    )

    const onMatchVersionChange = useCallback<React.ChangeEventHandler<HTMLInputElement>>(
        e => updateContext({ matchVersion: e.currentTarget.value }),
        [updateContext]
    )

    const onActionRequireVersionChange = useCallback<React.ChangeEventHandler<HTMLInputElement>>(
        e => {
            setActionRequireVersion(e.currentTarget.value)
            updateContext({ action: { requireVersion: e.currentTarget.value } })
        },
        [updateContext]
    )

    const onActionChange = useCallback<React.ChangeEventHandler<HTMLInputElement>>(
        e =>
            updateContext({
                action:
                    e.currentTarget.value === 'ban' ? e.currentTarget.value : { requireVersion: actionRequireVersion },
            }),
        [actionRequireVersion, updateContext]
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
                <div className="form-row">
                    <div className="col">
                        <label htmlFor="campaign-template-form__packageName">Dependency package name</label>
                        <input
                            type="text"
                            id="campaign-template-form__packageName"
                            className="form-control"
                            required={true}
                            minLength={1}
                            value={context.packageName || ''}
                            onChange={onPackageNameChange}
                            autoFocus={true}
                            disabled={disabled}
                        />
                        <p className="form-help text-muted small mb-0 mt-1">
                            Examples: <code className="border-bottom small mr-2">lodash</code>{' '}
                            <code className="border-bottom small mr-2">react</code>{' '}
                            <code className="border-bottom small mr-2">@babel/core</code>
                        </p>
                    </div>
                    <div className="col">
                        <label htmlFor="campaign-template-form__matchVersion">Version range</label>
                        <div className="input-group">
                            <div className="input-group-prepend">
                                <span className="input-group-text">@</span>
                            </div>
                            <input
                                type="text"
                                id="campaign-template-form__matchVersion"
                                className="form-control w-auto"
                                required={true}
                                minLength={1}
                                size={15}
                                value={context.matchVersion || ''}
                                onChange={onMatchVersionChange}
                                disabled={disabled}
                            />
                        </div>
                        <p className="form-help text-muted small mb-0 mt-1">
                            Supports{' '}
                            <a
                                href="https://docs.npmjs.com/misc/semver#ranges"
                                target="_blank"
                                rel="noopener noreferrer"
                            >
                                semver ranges
                            </a>
                            .
                        </p>
                    </div>
                </div>
            </div>
            <div className="form-group">
                <label htmlFor="campaign-template-form__action">Action</label>
                <ul className="list-unstyled">
                    <li>
                        <div className="form-check">
                            <input
                                type="radio"
                                id="campaign-template-form__action-minVersion-check"
                                className="form-check-input"
                                checked={context.action !== 'ban'}
                                value="requireVersion"
                                onChange={onActionChange}
                            />
                            <div>
                                <label
                                    className="form-check-label mb-3"
                                    htmlFor="campaign-template-form__action-minVersion-check"
                                >
                                    Upgrade to version{context.action === 'ban' ? '...' : ''}
                                </label>
                                {context.action !== 'ban' && (
                                    <>
                                        <input
                                            type="text"
                                            id="campaign-template-form__action-minVersion"
                                            className="form-control w-auto"
                                            required={true}
                                            minLength={1}
                                            placeholder=""
                                            size={30}
                                            value={context.action ? context.action.requireVersion : ''}
                                            onChange={onActionRequireVersionChange}
                                            disabled={disabled}
                                        />
                                        <p className="form-help text-muted small mb-0 mt-1">
                                            Supports{' '}
                                            <a
                                                href="https://docs.npmjs.com/misc/semver#ranges"
                                                target="_blank"
                                                rel="noopener noreferrer"
                                            >
                                                semver ranges
                                            </a>
                                            . Examples: <code className="border-bottom mr-3 small">&gt;=1.10.0</code>
                                            <code className="border-bottom mr-3 small">~0.2.2 || ^0.3.2</code>
                                        </p>
                                    </>
                                )}
                            </div>
                        </div>
                    </li>
                    <li>
                        <div className="form-check mt-2">
                            <input
                                type="radio"
                                id="campaign-template-form__action-ban-check"
                                className="form-check-input"
                                checked={context.action === 'ban'}
                                value="ban"
                                onChange={onActionChange}
                            />
                            <label className="form-check-label" htmlFor="campaign-template-form__action-ban-check">
                                Ban
                            </label>
                        </div>
                    </li>
                </ul>
            </div>
            <div className="form-group">
                <label>Options</label>
                <ul className="list-unstyled">
                    <li className="form-check">
                        <input
                            type="checkbox"
                            id="campaign-template-form__createChangesets"
                            className="form-check-input"
                            checked={!!context.createChangesets}
                            onChange={onCreateChangesetsChange}
                            disabled={disabled}
                        />
                        <label className="form-check-label" htmlFor="campaign-template-form__createChangesets">
                            Create changesets
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
    defaultWorkflow: {
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
                diagnostics: 'dependencyManagement.packageJsonDependency',
                codeActions: [{ command: 'dependencyManagement.packageJsonDependency.action' }],
            },
        ],
        behaviors: {
            edits: { command: 'changesets.byRepositoryAndBaseBranch' },
        },
    },
    workflowJSONSchema: {
        type: 'object',
        properties: {
            variables: {
                type: 'object',
                required: ['packageName'],
                properties: {
                    packageName: {
                        type: 'string',
                        description: 'The npm package name to operate on.',
                    },
                },
            },
        },
    },
    suggestTitle: (workflow): string | undefined => {
        if (!workflow || !workflow.variables) {
            return undefined
        }
        const { packageName, matchVersion, action } = workflow.variables
        const packageNameAndVersion = `${packageName || '<package>'}${matchVersion ? `@${matchVersion}` : ''}`
        const campaignName =
            action === 'ban'
                ? `Ban npm dependency ${packageNameAndVersion}`
                : `Upgrade npm dependency ${packageNameAndVersion} to ${
                      action && action.requireVersion ? action.requireVersion : '<version>'
                  }`
        return campaignName
    },
}
