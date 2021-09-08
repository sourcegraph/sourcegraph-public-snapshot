import * as H from 'history'
import React, { FunctionComponent } from 'react'

import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'

import { CodeIntelligenceConfigurationPolicyFields } from '../../../graphql-operations'

import { CodeIntelligencePolicyTable } from './CodeIntelligencePolicyTable'

export interface PoliciesListProps {
    policies?: CodeIntelligenceConfigurationPolicyFields[]
    deletePolicy?: (id: string, name: string) => Promise<void>
    disabled: boolean
    indexingEnabled: boolean
    buttonFragment?: JSX.Element
    history: H.History
}

export const PoliciesList: FunctionComponent<PoliciesListProps> = ({ policies, buttonFragment, ...props }) =>
    policies === undefined ? (
        <LoadingSpinner className="icon-inline" />
    ) : (
        <>
            {policies.length === 0 ? (
                <div>No policies have been defined.</div>
            ) : (
                <CodeIntelligencePolicyTable {...props} policies={policies} />
            )}
            {buttonFragment}
        </>
    )
