import NpmIcon from 'mdi-react/NpmIcon'
import React, { useCallback, useEffect, useState } from 'react'
import { CampaignTemplate, CampaignTemplateComponentContext } from '.'
import { RuleDefinition } from '../../../../rules/form/definition/RuleDefinitionFormControl'

interface Props extends CampaignTemplateComponentContext {}

const ALL_VERSION_RANGE = '*'

const PackageJsonDependencyCampaignTemplateForm: React.FunctionComponent<Props> = ({
    value,
    onChange,
    disabled,
    location,
}) => {
    const params = new URLSearchParams(location.search)
    const [packageName, setPackageName] = useState(params.get('packageName') || '')
    const onPackageNameChange = useCallback<React.ChangeEventHandler<HTMLInputElement>>(e => {
        setPackageName(e.currentTarget.value)
    }, [])

    const [versionRange, setVersionRange] = useState(params.get('versionRange') || ALL_VERSION_RANGE)
    const onVersionRangeChange = useCallback<React.ChangeEventHandler<HTMLInputElement>>(e => {
        setVersionRange(e.currentTarget.value)
    }, [])

    useEffect(() => {
        const packageNameOrPlaceholder = packageName || '<package>'
        onChange({
            ...value,
            name: `Deprecate ${packageNameOrPlaceholder}${
                versionRange && versionRange !== ALL_VERSION_RANGE ? `@${versionRange}` : ''
            } (npm)`,
            rules: [
                // TODO!(sqs): hack
                {
                    name: 'Find package.json dependencies entries',
                    // tslint:disable-next-line: no-object-literal-type-assertion
                    definition: JSON.stringify({
                        conditions: `file:(^|/)package\\.json$ '${JSON.stringify(packageNameOrPlaceholder)}'`,
                    } as RuleDefinition),
                },
            ],
        })
    }, [onChange, packageName, value, versionRange])

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
                    value={packageName}
                    onChange={onPackageNameChange}
                    autoFocus={true}
                    disabled={disabled}
                />
            </div>
            <div className="form-group">
                <label htmlFor="campaign-template-form__versionRange">Version range (to deprecate)</label>
                <input
                    type="text"
                    id="campaign-template-form__versionRange"
                    className="form-control"
                    placeholder="semver range (e.g., <1.2.3)"
                    value={versionRange}
                    onChange={onVersionRangeChange}
                    disabled={disabled}
                />
                <small className="form-help text-muted">
                    <a href="https://docs.npmjs.com/misc/semver#ranges">How to specify version ranges</a> (
                    <code>{ALL_VERSION_RANGE}</code> matches all versions)
                </small>
            </div>
        </>
    )
}

export const PackageJsonDependencyCampaignTemplate: CampaignTemplate = {
    id: 'packageJsonDependency',
    title: 'package.json dependency deprecation/ban',
    detail:
        'Deprecate or ban an npm/yarn dependency in package.json manifests, opening issues/changesets for all affected code owners.',
    icon: NpmIcon,
    renderForm: PackageJsonDependencyCampaignTemplateForm,
}
