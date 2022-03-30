import React, { FunctionComponent, useEffect, useMemo, useState } from 'react'

import { debounce } from 'lodash'

import { GitObjectType } from '../../../../graphql-operations'

import { GitObjectPreviewWrapper } from './GitObjectPreview'
import { Input } from "@sourcegraph/wildcard";

const DEBOUNCED_WAIT = 250

export interface ObjectsMatchingGitPatternProps {
    repoId?: string
    type: GitObjectType
    pattern: string
    setPattern: (pattern: string) => void
    disabled: boolean
}

export const ObjectsMatchingGitPattern: FunctionComponent<ObjectsMatchingGitPatternProps> = ({
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
                <div className="form-group">
                    <label htmlFor="pattern">Pattern</label>
                    <Input id="pattern" className="text-monospace" value={localPattern} onChange={({ target: { value } }) => {
                                                    setLocalPattern(value)
                                                    debouncedSetPattern(value)
                                                }} disabled={disabled} required={true}
                    />
                    <small className="form-text text-muted">Required.</small>
                </div>
            )}
            {repoId && <GitObjectPreviewWrapper repoId={repoId} type={type} pattern={pattern} />}
        </>
    )
}
