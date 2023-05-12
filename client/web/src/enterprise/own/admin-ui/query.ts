import { gql } from '@sourcegraph/http-client'

export const OwnSignalFragment = gql`
    fragment OwnSignalConfig on SignalConfiguration {
        name
        description
        isEnabled
        excludedRepoPatterns
    }
`

export const GET_OWN_JOB_CONFIGURATIONS = gql`
    query GetOwnSignalConfigurations {
        signalConfigurations {
            ... OwnSignalConfig
        }
    }
    ${OwnSignalFragment}
`

export const UPDATE_SIGNAL_CONFIGURATIONS = gql`
    mutation UpdateSignalConfigs($input:UpdateSignalConfigurationsInput!) {
        updateSignalConfigurations(input:$input) {
            isEnabled
            name
            description
            excludedRepoPatterns
        }
    }
`
