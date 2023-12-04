import { gql } from '@sourcegraph/http-client'

export const OWN_SIGNAL_FRAGMENT = gql`
    fragment OwnSignalConfig on OwnSignalConfiguration {
        name
        description
        isEnabled
        excludedRepoPatterns
    }
`

export const GET_OWN_JOB_CONFIGURATIONS = gql`
    query GetOwnSignalConfigurations {
        ownSignalConfigurations {
            ...OwnSignalConfig
        }
    }
    ${OWN_SIGNAL_FRAGMENT}
`

export const UPDATE_SIGNAL_CONFIGURATIONS = gql`
    mutation UpdateSignalConfigs($input: UpdateSignalConfigurationsInput!) {
        updateOwnSignalConfigurations(input: $input) {
            isEnabled
            name
            description
            excludedRepoPatterns
        }
    }
`

export const GET_INSTANCE_OWN_STATS = gql`
    query GetInstanceOwnStats {
        instanceOwnershipStats {
            totalFiles
            totalCodeownedFiles
            totalOwnedFiles
            totalAssignedOwnershipFiles
            updatedAt
        }
    }
`
