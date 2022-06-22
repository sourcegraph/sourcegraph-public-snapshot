import { FunctionComponent } from 'react'

import { CodeIntelligenceConfigurationPolicyFields } from '../../../../graphql-operations'

import { GitObjectTargetDescription } from './GitObjectTargetDescription'

export const LockfileIndexingPolicyDescription: FunctionComponent<
    React.PropsWithChildren<{ policy: CodeIntelligenceConfigurationPolicyFields }>
> = ({ policy }) =>
    policy.indexingEnabled ? (
        <>
            <strong>Lockfile indexing policy:</strong> index lockfiles <GitObjectTargetDescription policy={policy} />.
        </>
    ) : (
        <span className="text-muted">Lockfile-indexing disabled.</span>
    )
