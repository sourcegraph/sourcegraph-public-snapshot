import React, { useCallback } from 'react'

interface Props {
    value: string
    onChange: (value: string) => void
    disabled?: boolean
    className?: string
}

export const CampaignFormIncludeRepositoriesFormControl: React.FunctionComponent<Props> = ({
    value,
    onChange: parentOnChange,
    disabled,
    className = '',
}) => {
    const onChange = useCallback<React.ChangeEventHandler<HTMLInputElement>>(
        e => parentOnChange(e.currentTarget.value),
        [parentOnChange]
    )

    return (
        <div className={`form-group ${className}`}>
            <label htmlFor="campaign-form__includeRepositories">Include only specific repositories</label>
            <input
                type="text"
                id="campaign-form__includeRepositories"
                className="form-control"
                placeholder="Regular expression (e.g., myorg/)"
                value={value}
                onChange={onChange}
                disabled={disabled}
                list="campaign-form__includeRepositories-datalist"
            />
            <datalist id="campaign-form__includeRepositories-datalist">
                {/* tslint:disable-next-line: no-use-before-declare */}
                {SAMPLE_REPOSITORIES.map(repositoryName => (
                    <option key={repositoryName} value={repositoryName} />
                ))}
            </datalist>
        </div>
    )
}

const SAMPLE_REPOSITORIES = [
    'github.com/sd9/about',
    'github.com/sd9/codeintellify',
    'github.com/sd9/content-renderer',
    'github.com/sd9/go-diff',
    'github.com/sd9/gomemcache',
    'github.com/sd9/hackathon-starter',
    'github.com/sd9/lint',
    'github.com/sd9org/apiproxy',
    'github.com/sd9org/appmon',
    'github.com/sd9org/checkup',
    'github.com/sd9org/golang-lib',
    'github.com/sd9org/groupcache',
    'github.com/sd9org/java-maven-app',
    'github.com/sd9org/makex',
    'github.com/sd9org/react-router',
    'github.com/sd9/react-loading-spinner',
    'github.com/sd9/sourcegraph-lightstep',
    'github.com/sd9/TypeScriptSamples',
    'github.com/sourcegraph/about',
    'github.com/sourcegraph/codeintellify',
    'github.com/sourcegraph/go-diff',
    'github.com/sourcegraph/sourcegraph',
    'github.com/sourcegraph/sourcegraph-lightstep',
    'github.com/sqs/spans',
]
