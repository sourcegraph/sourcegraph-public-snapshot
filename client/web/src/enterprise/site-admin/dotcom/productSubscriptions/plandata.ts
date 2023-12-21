export interface Tag {
    name: string
    tagValue: string
    description?: string
}

export interface Plan {
    name: string
    label: string
    deprecated?: boolean
    cloudOnlyPlan?: boolean
    additionalTags?: Tag[]
}

export const TAG_BATCH_CHANGES: Tag = {
    name: 'Batch Changes',
    tagValue: 'batch-changes',
    description: 'Allows to use the Batch Changes feature beyond the 10 changeset limit',
}

export const TAG_CODE_INSIGHTS: Tag = {
    name: 'Code Insights',
    tagValue: 'code-insights',
    description: 'Allows to use the Code Insights feature',
}

export const TAG_TRIAL: Tag = {
    name: 'Trial',
    tagValue: 'trial',
    description: 'Whether the license has been issued as part of a trial',
}

export const TAG_AIR_GAPPED: Tag = {
    name: 'Allow Air Gapped Usage',
    tagValue: 'allow-air-gapped',
    description: 'Allows offline usage of Sourcegraph with no remote license key validation and telemetry off',
}

export const TAG_TRUEUP: Tag = {
    name: 'Allows usage of the true up billing model',
    tagValue: 'true-up',
    description: 'Allows the instance to go over the user hard cap and be billed based on actual usage',
}

export const DEPRECATED_TAGS: Tag[] = [
    {
        name: 'AllowAnonymousUsage',
        tagValue: 'allow-anonymous-usage',
        description: 'Denotes licenses that allow anonymous usage, a.k.a public access to the instance',
    },
    {
        name: 'PrivateExtensionRegistry',
        tagValue: 'private-extension-registry',
        description:
            'Whether publishing extensions to this Sourcegraph instance has been purchased. If not, then extensions must be published to Sourcegraph.com. All instances may use extensions published to Sourcegraph.com',
    },
    {
        name: 'RemoteExtensionsAllowDisallow',
        tagValue: 'remote-extensions-allow-disallow',
        description:
            'Whether explicitly specify a list of allowed remote extensions and prevent any other remote extensions from being used has been purchased. It does not apply to locally published extensions',
    },
    {
        name: 'Branding',
        tagValue: 'branding',
        description: 'Whether custom branding of this Sourcegraph instance has been purchased',
    },
    {
        name: 'Campaigns',
        tagValue: 'campaigns',
        description: 'Whether campaigns (now: batch changes) on this Sourcegraph instance has been purchased',
    },
    {
        name: 'Monitoring',
        tagValue: 'monitoring',
        description: 'Whether monitoring on this Sourcegraph instance has been purchased',
    },
    {
        name: 'BackupAndRestore',
        tagValue: 'backup-and-restore',
        description: 'Whether builtin backup and restore on this Sourcegraph instance has been purchased',
    },
    {
        name: 'MAU',
        tagValue: 'mau',
        description: 'The customer is on a MAU pricing model',
    },
]

export const ALL_PLANS: Plan[] = [
    {
        name: 'Code Search',
        label: 'code-search',
        additionalTags: [TAG_TRIAL, TAG_BATCH_CHANGES, TAG_CODE_INSIGHTS, TAG_AIR_GAPPED],
    },
    {
        name: 'Code Intelligence Platform',
        label: 'cip',
        additionalTags: [TAG_TRIAL, TAG_AIR_GAPPED],
    },
    {
        name: 'Cody Only Enterprise',
        label: 'cody-only-enterprise',
        additionalTags: [TAG_TRIAL, TAG_AIR_GAPPED],
    },
    {
        name: 'Cody Search Enterprise Only',
        label: 'cody-search-enterprise-only',
        additionalTags: [TAG_TRIAL, TAG_AIR_GAPPED],
    },
    {
        name: 'Code AI Enterprise',
        label: 'code-ai-enterprise',
        additionalTags: [TAG_TRIAL, TAG_AIR_GAPPED, TAG_TRUEUP],
    },

    // Old plans at the bottom for convenience:
    {
        name: 'old-starter-0',
        label: 'old-starter-0',
        deprecated: true,
    },
    {
        name: 'old-enterprise-0',
        label: 'old-enterprise-0',
        deprecated: true,
    },
    {
        name: 'team-0',
        label: 'team-0',
        deprecated: true,
    },
    {
        name: 'enterprise-0',
        label: 'enterprise-0',
        deprecated: true,
    },
    {
        name: 'enterprise-1',
        label: 'enterprise-1',
        deprecated: true,
    },
    {
        name: 'enterprise-extension',
        label: 'enterprise-extension',
        deprecated: true,
    },
    {
        name: 'free-0',
        label: 'free-0',
        deprecated: true,
    },
    {
        name: 'free-1',
        label: 'free-1',
        deprecated: true,
    },
    {
        name: 'enterprise-air-gap-0',
        label: 'enterprise-air-gap-0',
        deprecated: true,
    },
]
