import * as H from 'history'
import React, { FunctionComponent } from 'react'

import { ErrorAlert } from '@sourcegraph/web/src/components/alerts'
import { Container } from '@sourcegraph/wildcard'

import { CodeIntelligenceConfigurationPolicyFields } from '../../../graphql-operations'

import { PoliciesList } from './PoliciesList'
import { PolicyListActions } from './PolicyListActions'

export interface GlobalPoliciesProps {
    repo?: { id: string }
    disabled: boolean
    deleting: boolean
    globalPolicies?: CodeIntelligenceConfigurationPolicyFields[]
    deleteGlobalPolicy: (id: string, name: string) => Promise<void>
    deleteError?: Error
    indexingEnabled: boolean
    history: H.History
}

export const GlobalPolicies: FunctionComponent<GlobalPoliciesProps> = ({
    repo,
    disabled,
    deleting,
    globalPolicies,
    deleteGlobalPolicy,
    deleteError,
    indexingEnabled,
    history,
}) => (
    <Container>
        <h3>Global policies</h3>

        {repo === undefined && deleteError && (
            <ErrorAlert prefix="Error deleting configuration policy" error={deleteError} />
        )}

        <PoliciesList
            policies={globalPolicies}
            deletePolicy={repo ? undefined : deleteGlobalPolicy}
            disabled={disabled}
            indexingEnabled={indexingEnabled}
            buttonFragment={
                repo === undefined ? (
                    <PolicyListActions disabled={disabled} deleting={deleting} history={history} />
                ) : undefined
            }
            history={history}
        />
    </Container>
)
