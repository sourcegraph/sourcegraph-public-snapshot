export const LogsPageTabs = {
    COMMANDS: 0,
    SYNCLOGS: 1,
} as const

export enum CodeHostType {
    GITHUB = 'github',
    BITBUCKETCLOUD = 'bitbucketCloud',
    BITBUCKETSERVER = 'bitbucketServer',
    GITLAB = 'gitlab',
    GITOLITE = 'gitolite',
    AWSCODECOMMIT = 'awsCodeCommit',
    AZUREDEVOPS = 'azureDevOps',
    OTHER = 'other',
}
