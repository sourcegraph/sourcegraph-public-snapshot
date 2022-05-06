import { FunctionComponent } from 'react'

import { CodeIntelligenceConfigurationPolicyFields } from '../../../../graphql-operations'
import { formatDurationValue } from '../shared'

import { GitObjectTargetDescription } from './GitObjectTargetDescription'

export const IndexingPolicyDescription: FunctionComponent<
    React.PropsWithChildren<{ policy: CodeIntelligenceConfigurationPolicyFields }>
> = ({ policy }) =>
    policy.indexingEnabled ? (
        <>
            <strong>Indexing policy:</strong> Auto-index <GitObjectTargetDescription policy={policy} />
            {policy.indexCommitMaxAgeHours && (
                <> if the target commit is no older than {formatDurationValue(policy.indexCommitMaxAgeHours)}</>
            )}
            .
        </>
    ) : (
        <span className="text-muted">Auto-indexing disabled.</span>
    )
