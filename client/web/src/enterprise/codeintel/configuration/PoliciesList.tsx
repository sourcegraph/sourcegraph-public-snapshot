import * as H from 'history'
import React, { FunctionComponent } from 'react'

import { CodeIntelligenceConfigurationPolicyFields } from '../../../graphql-operations'

import { CodeIntelligencePolicyTable } from './CodeIntelligencePolicyTable'
import { DeletePolicyResult } from './usePoliciesConfigurations'

export interface PoliciesListProps {
    policies: CodeIntelligenceConfigurationPolicyFields[]
    onDeletePolicy?: (id: string, name: string) => DeletePolicyResult
    disabled: boolean
    indexingEnabled: boolean
    buttonFragment?: JSX.Element
    history: H.History
}

export const PoliciesList: FunctionComponent<PoliciesListProps> = ({
    policies,
    buttonFragment,
    onDeletePolicy,
    ...props
}) => (
    <>
        {policies.length === 0 ? (
            <div>No policies have been defined.</div>
        ) : (
            <CodeIntelligencePolicyTable {...props} policies={policies} onDeletePolicy={onDeletePolicy} />
        )}
        {buttonFragment}
    </>
)
