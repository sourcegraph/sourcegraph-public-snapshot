import * as H from 'history'
import React, { FunctionComponent } from 'react'

import { CodeIntelligenceConfigurationPolicyFields } from '../../../graphql-operations'

import { CodeIntelligencePolicyTable } from './CodeIntelligencePolicyTable'

export interface PoliciesListProps {
    policies: CodeIntelligenceConfigurationPolicyFields[]
    onDeletePolicy?: (id: string, name: string) => Promise<void>
    disabled: boolean
    indexingEnabled: boolean
    history: H.History
}

export const PoliciesList: FunctionComponent<PoliciesListProps> = ({ policies, onDeletePolicy, ...props }) => (
    <>
        {policies.length === 0 ? (
            <div>No policies have been defined.</div>
        ) : (
            <CodeIntelligencePolicyTable {...props} policies={policies} onDeletePolicy={onDeletePolicy} />
        )}
    </>
)
