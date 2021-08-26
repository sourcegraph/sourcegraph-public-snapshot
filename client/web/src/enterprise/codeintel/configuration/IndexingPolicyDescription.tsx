import React, { FunctionComponent } from 'react'

import { CodeIntelligenceConfigurationPolicyFields } from '../../../graphql-operations'

import { GitObjectTargetDescription } from './GitObjectTargetDescription'
import { formatDurationValue } from './shared'

export const IndexingPolicyDescription: FunctionComponent<{ policy: CodeIntelligenceConfigurationPolicyFields }> = ({
    policy,
}) =>
    policy.indexingEnabled ? (
        <>
            <strong>Indexing policy:</strong> Auto-index <GitObjectTargetDescription policy={policy} />
            {policy.indexCommitMaxAgeHours !== 0 && (
                <> if the target commit is no older than {formatDurationValue(policy.indexCommitMaxAgeHours)}</>
            )}
            .
        </>
    ) : (
        <span className="text-muted">Auto-indexing disabled.</span>
    )
