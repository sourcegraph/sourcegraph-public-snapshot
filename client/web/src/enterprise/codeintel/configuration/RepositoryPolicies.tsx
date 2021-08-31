import * as H from 'history'
import React, { FunctionComponent } from 'react'

import { ErrorAlert } from '@sourcegraph/web/src/components/alerts'
import { Container } from '@sourcegraph/wildcard'

import { CodeIntelligenceConfigurationPolicyFields } from '../../../graphql-operations'

import { PoliciesList } from './PoliciesList'
import { PolicyListActions } from './PolicyListActions'

export interface RepositoryPoliciesProps {
    disabled: boolean
    deleting: boolean
    policies?: CodeIntelligenceConfigurationPolicyFields[]
    deletePolicy: (id: string, name: string) => Promise<void>
    deleteError?: Error
    indexingEnabled: boolean
    history: H.History
}

export const RepositoryPolicies: FunctionComponent<RepositoryPoliciesProps> = ({
    disabled,
    deleting,
    policies,
    deletePolicy,
    deleteError,
    indexingEnabled,
    history,
}) => (
    <Container>
        <h3>Repository-specific policies</h3>

        {deleteError && <ErrorAlert prefix="Error deleting configuration policy" error={deleteError} />}

        <PoliciesList
            policies={policies}
            deletePolicy={deletePolicy}
            disabled={disabled}
            indexingEnabled={indexingEnabled}
            buttonFragment={<PolicyListActions disabled={disabled} deleting={deleting} history={history} />}
            history={history}
        />
    </Container>
)
