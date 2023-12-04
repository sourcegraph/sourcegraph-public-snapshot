import type { CodeIntelligenceConfigurationPolicyFields } from '../../../graphql-operations'

export const defaultDurationValues = [
    // These values match the output of date-fns#formatDuration
    { value: 168, displayText: '7 days' },
    { value: 744, displayText: '1 month' }, // 31 days
    { value: 2160, displayText: '3 months' }, // 90 days
    { value: 8760, displayText: '1 year' }, // 365 days
    { value: 43824, displayText: '5 years' }, // 4 years + 1 leap day
]

export const formatDurationValue = (value: number): string => {
    const match = defaultDurationValues.find(candidate => candidate.value === value)
    if (!match) {
        return `${value} hours`
    }

    return match.displayText
}

export const hasGlobalPolicyViolation = (policy: CodeIntelligenceConfigurationPolicyFields): boolean => {
    // If there are no repo patterns, it is assumed that the policy targets all repos.
    const repoPatterns = policy.repositoryPatterns || []

    return (
        // User has enabled auto indexing for a policy
        policy.indexingEnabled &&
        // Policy isn't targeted at a specific repository
        !policy.repository &&
        // Policy does not have a targeted repository pattern.
        // TODO(#47432): This is flaky as repoPatterns can match all repositories (e.g. '*'). We should return a flag that indicates if this has happened.
        repoPatterns.length === 0
    )
}
