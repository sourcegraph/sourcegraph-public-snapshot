import { FunctionComponent } from 'react'

import { CodeIntelligenceConfigurationPolicyFields } from '../../../../graphql-operations'
import { formatDurationValue } from '../shared'

import { GitObjectTargetDescription } from './GitObjectTargetDescription'

export const RetentionPolicyDescription: FunctionComponent<
    React.PropsWithChildren<{ policy: CodeIntelligenceConfigurationPolicyFields }>
> = ({ policy }) =>
    policy.retentionEnabled ? (
        <>
            <strong>Retention policy:</strong>{' '}
            <span>
                Retain uploads used to resolve code intelligence queries for{' '}
                <GitObjectTargetDescription policy={policy} />
                {policy.retentionDurationHours && (
                    <> for at least {formatDurationValue(policy.retentionDurationHours)} after upload</>
                )}
                .
            </span>
        </>
    ) : (
        <span className="text-muted">Data retention disabled.</span>
    )
