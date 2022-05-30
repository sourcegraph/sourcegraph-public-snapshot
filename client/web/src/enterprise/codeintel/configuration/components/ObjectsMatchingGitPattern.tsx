import { FunctionComponent, useEffect, useMemo, useState } from 'react'

import { debounce } from 'lodash'

import { Input } from '@sourcegraph/wildcard'

import { GitObjectType } from '../../../../graphql-operations'

import { GitObjectPreviewWrapper } from './GitObjectPreview'

const DEBOUNCED_WAIT = 250

export interface ObjectsMatchingGitPatternProps {
    repoId?: string
    type: GitObjectType
    pattern: string
    setPattern: (pattern: string) => void
    disabled: boolean
}

export const ObjectsMatchingGitPattern: FunctionComponent<React.PropsWithChildren<ObjectsMatchingGitPatternProps>> = ({
    repoId,
    type,
    pattern,
    setPattern,
    disabled,
}) => {
    const [localPattern, setLocalPattern] = useState('')
    useEffect(() => setLocalPattern(pattern), [pattern])

    const debouncedSetPattern = useMemo(() => debounce(value => setPattern(value), DEBOUNCED_WAIT), [setPattern])

    return (
        <>
            {type !== GitObjectType.GIT_COMMIT && (
                <Input
                    id="pattern"
                    label="Pattern"
                    className="form-group"
                    inputClassName="text-monospace"
                    value={localPattern}
                    onChange={({ target: { value } }) => {
                        setLocalPattern(value)
                        debouncedSetPattern(value)
                    }}
                    disabled={disabled}
                    required={true}
                    message="Required"
                />
            )}
            {repoId && <GitObjectPreviewWrapper repoId={repoId} type={type} pattern={pattern} />}
        </>
    )
}
