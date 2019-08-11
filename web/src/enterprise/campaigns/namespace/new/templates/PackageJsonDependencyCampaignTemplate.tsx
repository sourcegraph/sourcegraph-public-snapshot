import NpmIcon from 'mdi-react/NpmIcon'
import React, { useCallback, useEffect, useState } from 'react'
import { CampaignTemplate, CampaignTemplateComponentContext } from '.'
import { pluralize } from '../../../../../../../shared/src/util/strings'
import { RuleDefinition } from '../../../../rules/types'

interface Props extends CampaignTemplateComponentContext {}

const ALL_VERSION_RANGE = '*'

const PackageJsonDependencyCampaignTemplateForm: React.FunctionComponent<Props> = ({
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
            name: `Deprecate ${packageNameOrPlaceholder}${
                versionRange && versionRange !== ALL_VERSION_RANGE ? `@${versionRange}` : ''
            } (npm)`,
            rules: packageName
                ? [
                      // TODO!(sqs): hack
                      {
                          name: 'Find package.json dependencies entries',
                          // tslint:disable-next-line: no-object-literal-type-assertion
                          definition: JSON.stringify({
                              type: 'DiagnosticRule',
                              query: {
                                  type: 'packageJsonDependency',
                                  tag: [packageNameOrPlaceholder],
                              },
                              action: 'packageJsonDependency.remove',
                          } as RuleDefinition),
                      },
                  ]
                : [],
        })
    }, [onChange, packageName, versionRange])

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
                    list="campaign-template-form__packageName-datalist"
                />
                <datalist id="campaign-template-form__packageName-datalist">
                    {/* tslint:disable-next-line: no-use-before-declare */}
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
                    value={versionRange}
                    onChange={onVersionRangeChange}
                    disabled={disabled}
                />
                <small className="form-help text-muted">
                    <a href="https://docs.npmjs.com/misc/semver#ranges" target="_blank">
                        How to specify version ranges
                    </a>{' '}
                    (<code>{ALL_VERSION_RANGE}</code> matches all versions)
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

const SAMPLE_PACKAGE_NAMES: { packageName: string; count: number }[] = [
    { packageName: 'typescript', count: 5 },
    { packageName: 'react', count: 3 },
    { packageName: 'lodash', count: 3 },
    { packageName: 'mdi-react', count: 3 },
    { packageName: 'glob', count: 2 },
    { packageName: '@sourcegraph/codeintellify', count: 1 },
]
