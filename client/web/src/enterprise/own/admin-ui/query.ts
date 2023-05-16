import { gql } from '@sourcegraph/http-client'

export const OWN_SIGNAL_FRAGMENT = gql`
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
    ${OWN_SIGNAL_FRAGMENT}
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
