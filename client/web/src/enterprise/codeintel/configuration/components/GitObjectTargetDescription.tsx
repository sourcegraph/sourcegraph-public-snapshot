import { FunctionComponent } from 'react'

import { GitObjectType } from '@sourcegraph/shared/src/schema'

import { CodeIntelligenceConfigurationPolicyFields } from '../../../../graphql-operations'

export const GitObjectTargetDescription: FunctionComponent<
    React.PropsWithChildren<{ policy: CodeIntelligenceConfigurationPolicyFields }>
> = ({ policy }) =>
    policy.type === GitObjectType.GIT_COMMIT ? (
        <>the matching commit</>
    ) : policy.type === GitObjectType.GIT_TAG ? (
        <>the matching tags</>
    ) : policy.type === GitObjectType.GIT_TREE ? (
        !policy.retainIntermediateCommits ? (
            <>the tip of the matching branches</>
        ) : (
            <>any commit on the matching branches</>
        )
    ) : (
        <></>
    )
