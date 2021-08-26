import React, { FunctionComponent } from 'react'

import { CodeIntelligenceConfigurationPolicyFields } from '../../../graphql-operations'

import { GitObjectTargetDescription } from './GitObjectTargetDescription'
import { formatDurationValue } from './shared'

export const RetentionPolicyDescription: FunctionComponent<{ policy: CodeIntelligenceConfigurationPolicyFields }> = ({
    policy,
}) =>
    policy.retentionEnabled ? (
        <>
            <strong>Retention policy:</strong>{' '}
            <span>
                Retain uploads used to resolve code intelligence queries for{' '}
                <GitObjectTargetDescription policy={policy} />
                {policy.retentionDurationHours !== 0 && (
                    <> for at least {formatDurationValue(policy.retentionDurationHours)} after upload</>
                )}
                .
            </span>
        </>
    ) : (
        <span className="text-muted">Data retention disabled.</span>
    )
