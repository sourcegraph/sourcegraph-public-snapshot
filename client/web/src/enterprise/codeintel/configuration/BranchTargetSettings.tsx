import { debounce } from 'lodash'
import React, { FunctionComponent, useState } from 'react'

import { CodeIntelligenceConfigurationPolicyFields, GitObjectType } from '../../../graphql-operations'

import {
    repoName as defaultRepoName,
    searchGitBranches as defaultSearchGitBranches,
    searchGitTags as defaultSearchGitTags,
} from './backend'
import { GitObjectPreview } from './GitObjectPreview'

export interface BranchTargetSettingsProps {
    repoId?: string
    policy: CodeIntelligenceConfigurationPolicyFields
    setPolicy: (policy: CodeIntelligenceConfigurationPolicyFields) => void
    repoName: typeof defaultRepoName
    searchGitBranches: typeof defaultSearchGitBranches
    searchGitTags: typeof defaultSearchGitTags
    disabled: boolean
}

const GIT_OBJECT_PREVIEW_DEBOUNCE_TIMEOUT = 300

export const BranchTargetSettings: FunctionComponent<BranchTargetSettingsProps> = ({
    repoId,
    policy,
    setPolicy,
    repoName,
    searchGitBranches,
    searchGitTags,
    disabled = false,
}) => {
    const [debouncedPattern, setDebouncedPattern] = useState(policy.pattern)
    const setPattern = debounce(value => setDebouncedPattern(value), GIT_OBJECT_PREVIEW_DEBOUNCE_TIMEOUT)

    return (
        <>
            <div className="form-group">
                <label htmlFor="name">Name</label>
                <input
                    id="name"
                    type="text"
                    className="form-control"
                    value={policy.name}
                    onChange={event => setPolicy({ ...policy, name: event.target.value })}
                    disabled={disabled}
                />
            </div>

            <div className="form-group">
                <label htmlFor="type">Type</label>
                <select
                    id="type"
                    className="form-control"
                    value={policy.type}
                    onChange={event =>
                        setPolicy({
                            ...policy,
                            type: event.target.value as GitObjectType,
                            ...(event.target.value !== GitObjectType.GIT_TREE
                                ? {
                                      retainIntermediateCommits: false,
                                      indexIntermediateCommits: false,
                                  }
                                : {}),
                        })
                    }
                    disabled={disabled}
                >
                    <option value="">Select Git object type</option>
                    {repoId && <option value={GitObjectType.GIT_COMMIT}>Commit</option>}
                    <option value={GitObjectType.GIT_TAG}>Tag</option>
                    <option value={GitObjectType.GIT_TREE}>Branch</option>
                </select>
            </div>
            <div className="form-group">
                <label htmlFor="pattern">Pattern</label>
                <input
                    id="pattern"
                    type="text"
                    className="form-control text-monospace"
                    value={policy.pattern}
                    onChange={event => {
                        setPolicy({ ...policy, pattern: event.target.value })
                        setPattern(event.target.value)
                    }}
                    disabled={disabled}
                />
            </div>

            {repoId && (
                <GitObjectPreview
                    pattern={debouncedPattern}
                    repoId={repoId}
                    type={policy.type}
                    repoName={repoName}
                    searchGitTags={searchGitTags}
                    searchGitBranches={searchGitBranches}
                />
            )}
        </>
    )
}
