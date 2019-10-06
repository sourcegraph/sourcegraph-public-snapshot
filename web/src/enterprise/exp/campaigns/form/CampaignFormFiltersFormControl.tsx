import React, { useCallback } from 'react'

interface Props {
    value: string
    onChange: (value: string) => void
    disabled?: boolean
    className?: string
}

export const CampaignFormFiltersFormControl: React.FunctionComponent<Props> = ({
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
            <label htmlFor="campaign-form__filters">Filters</label>
            <input
                type="text"
                id="campaign-form__filters"
                className="form-control"
                placeholder="repo:myrepo owner:alice"
                value={value}
                onChange={onChange}
                disabled={disabled}
                list="campaign-form__filters-datalist"
            />
            <datalist id="campaign-form__filters-datalist">
                {/* tslint:disable-next-line: no-use-before-declare */}
                {SAMPLE_FILTERS.map(repositoryName => (
                    <option key={repositoryName} value={repositoryName} />
                ))}
            </datalist>
        </div>
    )
}

const SAMPLE_FILTERS = [
    'repo:github.com/sd9/about',
    'repo:github.com/sd9/codeintellify',
    'repo:github.com/sd9/content-renderer',
    'repo:github.com/sd9/go-diff',
    'repo:github.com/sd9/gomemcache',
    'repo:github.com/sd9/hackathon-starter',
    'repo:github.com/sd9/lint',
    'repo:github.com/sd9org/apiproxy',
    'repo:github.com/sd9org/appmon',
    'repo:github.com/sd9org/checkup',
    'repo:github.com/sd9org/golang-lib',
    'repo:github.com/sd9org/groupcache',
    'repo:github.com/sd9org/java-maven-app',
    'repo:github.com/sd9org/makex',
    'repo:github.com/sd9org/react-router',
    'repo:github.com/sd9/react-loading-spinner',
    'repo:github.com/sd9/sourcegraph-lightstep',
    'repo:github.com/sd9/TypeScriptSamples',
    'repo:github.com/sourcegraph/about',
    'repo:github.com/sourcegraph/codeintellify',
    'repo:github.com/sourcegraph/go-diff',
    'repo:github.com/sourcegraph/sourcegraph',
    'repo:github.com/sourcegraph/sourcegraph-lightstep',
    'repo:github.com/sqs/spans',
]
