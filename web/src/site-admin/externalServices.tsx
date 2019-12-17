import { Edit, FormattingOptions, JSONPath } from '@sqs/jsonc-parser'
import { setProperty } from '@sqs/jsonc-parser/lib/edit'
import { flatMap, map } from 'lodash'
import AmazonIcon from 'mdi-react/AmazonIcon'
import BitbucketIcon from 'mdi-react/BitbucketIcon'
import GithubCircleIcon from 'mdi-react/GithubCircleIcon'
import GitIcon from 'mdi-react/GitIcon'
import GitLabIcon from 'mdi-react/GitlabIcon'
import React from 'react'
import { Link } from 'react-router-dom'
import awsCodeCommitSchemaJSON from '../../../schema/aws_codecommit.schema.json'
import bitbucketCloudSchemaJSON from '../../../schema/bitbucket_cloud.schema.json'
import bitbucketServerSchemaJSON from '../../../schema/bitbucket_server.schema.json'
import githubSchemaJSON from '../../../schema/github.schema.json'
import gitlabSchemaJSON from '../../../schema/gitlab.schema.json'
import gitoliteSchemaJSON from '../../../schema/gitolite.schema.json'
import otherExternalServiceSchemaJSON from '../../../schema/other_external_service.schema.json'
import phabricatorSchemaJSON from '../../../schema/phabricator.schema.json'
import { PhabricatorIcon } from '../../../shared/src/components/icons'
import * as GQL from '../../../shared/src/graphql/schema'
import { EditorAction } from './configHelpers.js'

/**
 * Metadata associated with a given external service.
 */
export interface ExternalServiceKindMetadata {
    /**
     * Title to show in the external service "button"
     */
    title: string

    /**
     * Icon to show in the external service "button"
     */
    icon: React.ComponentType<{ className?: string }>

    /**
     * A short description that will appear in the external service "button" under the title
     */
    shortDescription: string

    /**
     * A long description that will appear on the add / edit page
     */
    longDescription?: JSX.Element | string

    /**
     * The JSON schema of the external service configuration
     */
    jsonSchema: { $id: string }

    /**
     * Quick configure editor actions
     */
    editorActions?: EditorAction[]

    /**
     * Default display name
     */
    defaultDisplayName: string

    /**
     * Default external service configuration
     */
    defaultConfig: string
}

const defaultFormattingOptions: FormattingOptions = {
    eol: '\n',
    insertSpaces: true,
    tabSize: 2,
}

/**
 * editWithComment returns a Monaco edit action that sets the value of a JSON field with a
 * "//" comment annotating the field. The comment is inserted wherever
 * `"COMMENT_SENTINEL": true` appears in the JSON.
 */
function editWithComment(config: string, path: JSONPath, value: any, comment: string): Edit {
    const edit = setProperty(config, path, value, defaultFormattingOptions)[0]
    edit.content = edit.content.replace('"COMMENT_SENTINEL": true', comment)
    return edit
}

const editorActionComments = {
    enablePermissions:
        '// Prerequisite: you must configure GitHub as an OAuth auth provider in the site config (https://docs.sourcegraph.com/admin/auth#github). Otherwise, access to all repositories will be disallowed.',
    enforcePermissionsOAuth: `// Prerequisite: you must first update the site configuration to
    // include GitLab OAuth as an auth provider.
    // See https://docs.sourcegraph.com/admin/auth#gitlab for instructions.`,
    enforcePermissionsSSO: `// Prerequisite: You will need a sudo-level access token. If you can configure
    // GitLab as an OAuth identity provider for Sourcegraph, we recommend that
    // option instead.
    //
    // 1. Ensure the personal access token in this config has admin privileges
    //    (https://docs.gitlab.com/ee/api/#sudo).
    // 2. Update the site configuration to include the SSO auth provider for GitLab
    //    (https://docs.sourcegraph.com/admin/auth).
    // 3. Update the fields below to match the properties of this auth provider
    //    (https://docs.sourcegraph.com/admin/repo/permissions#sudo-access-token).`,
}

export const GITHUB_EXTERNAL_SERVICE: ExternalServiceKindMetadata = {
    title: 'GitHub repositories',
    icon: GithubCircleIcon,
    jsonSchema: githubSchemaJSON,
    editorActions: [
        {
            id: 'setAccessToken',
            label: 'Set access token',
            run: config => {
                const value = '<access token>'
                const edits = setProperty(config, ['token'], value, defaultFormattingOptions)
                return { edits, selectText: '<access token>' }
            },
        },
        {
            id: 'addOrgRepo',
            label: 'Add organization repositories',
            run: config => {
                const value = '<organization name>'
                const edits = setProperty(config, ['orgs', -1], value, defaultFormattingOptions)
                return { edits, selectText: '<organization name>' }
            },
        },
        {
            id: 'addRepo',
            label: 'Add a repository',
            run: config => {
                const value = '<owner>/<repository>'
                const edits = setProperty(config, ['repos', -1], value, defaultFormattingOptions)
                return { edits, selectText: '<owner>/<repository>' }
            },
        },
        {
            id: 'excludeRepo',
            label: 'Exclude a repository',
            run: config => {
                const value = { name: '<owner>/<repository>' }
                const edits = setProperty(config, ['exclude', -1], value, defaultFormattingOptions)
                return { edits, selectText: '{"name": "<owner>/<repository>"}' }
            },
        },
        {
            id: 'addSearchQueryRepos',
            label: 'Add repositories matching search query',
            run: config => {
                const value = '<search query>'
                const edits = setProperty(config, ['repositoryQuery', -1], value, defaultFormattingOptions)
                return { edits, selectText: '<search query>' }
            },
        },
        {
            id: 'enablePermissions',
            label: 'Enforce permissions',
            run: config => {
                const value = {
                    COMMENT_SENTINEL: true,
                }
                const comment = editorActionComments.enablePermissions
                const edit = editWithComment(config, ['authorization'], value, comment)
                return { edits: [edit], selectText: comment }
            },
        },
    ],
    shortDescription: 'Add GitHub.com repositories',
    longDescription: (
        <span>
            Configure by using the <strong>Quick configure</strong> buttons, or manually edit the JSON configuration.{' '}
            <Link target="_blank" to="/help/admin/external_service/github#configuration">
                Read the docs
            </Link>{' '}
            for more info about each field.
        </span>
    ),
    defaultDisplayName: 'GitHub',
    defaultConfig: `// Use Ctrl+Space for completion, and hover over JSON properties for documentation.
// GitHub external service docs: https://docs.sourcegraph.com/admin/external_service/github
{
  "url": "https://github.com",

  // token: GitHub API access token. Visit https://github.com/settings/tokens/new?scopes=repo&description=Sourcegraph
  // to create a token with access to public and private repositories
  "token": "<access token>",

  // SELECTING REPOSITORIES
  //
  // There are 4 fields used to select repositories for searching and code intel:
  //  - repositoryQuery (required)
  //  - orgs
  //  - repos
  //  - exclude

  // repositoryQuery: List of strings, either a special keyword ("none" or "affiliated"), or
  // GitHub search qualifiers, e.g. "archived:false"
  //
  // For getting started, use either:
  //  - "org:<name>" // (e.g. "org:sourcegraph") all repositories belonging to the organization
  // or
  //  - "affiliated" // all repositories affiliated (accessible) by the token's owner
  //
  // Additional query strings can be added to refine results:
  //  - "archived:false fork:no created:>=2016" // use of multiple search qualifiers
  //  - "user:docker repo:kubernetes/kubernetes" // fetch repositories outside of the user/org account
  //
  // See https://help.github.com/en/articles/searching-for-repositories for the list of search qualifiers.
  "repositoryQuery": [
  // "org:<name>" // set this to "none" to disable querying
  ],

  // orgs: List of organizations whose repositories should be selected
  // "orgs": [
  //   "<org name>"
  // ],

  // repos: Explicit list of repositories to select
  // "repos": [
  //   "<owner>/<repository>"
  // ],

  // exclude: Repositories to exclude (overrides repositories from repositoryQuery, orgs, and repos)
  // "exclude": [
  //   {
  //     "name": "<owner>/<repository>"
  //   }
  // ]
}`,
}

export const ALL_EXTERNAL_SERVICES: Record<GQL.ExternalServiceKind, ExternalServiceKindMetadata> = {
    [GQL.ExternalServiceKind.GITHUB]: GITHUB_EXTERNAL_SERVICE,
    [GQL.ExternalServiceKind.AWSCODECOMMIT]: {
        title: 'AWS CodeCommit repositories',
        icon: AmazonIcon,
        shortDescription: 'Add AWS CodeCommit repositories.',
        jsonSchema: awsCodeCommitSchemaJSON,
        defaultDisplayName: 'AWS CodeCommit',
        defaultConfig: `// Use Ctrl+Space for completion, and hover over JSON properties for documentation.
// AWS CodeCommit external service docs: https://docs.sourcegraph.com/admin/external_service/aws_codecommit#configuration
{
  "accessKeyID": "<access key id>",
  "secretAccessKey": "<secret access key>",
  "region": "<region>",

  // Git credentials for cloning an AWS CodeCommit repository over https
  // See IAM Code Commit auth docs: https://docs.aws.amazon.com/codecommit/latest/userguide/setting-up-gc.html
  "gitCredentials": {
    "username": "<username>",
    "password": "<password>"
  },

  // Repositories to exclude by name ({"name": "git-codecommit.us-west-1.amazonaws.com/repo-name"})
  // or by ARN ({"id": "arn:aws:codecommit:us-west-1:999999999999:name"})
  // "exclude": [
  //   {
  //     "name": "mono-repo"
  //   }
  // ]
}`,
        editorActions: [
            {
                id: 'setAccessKeyID',
                label: 'Set access key ID',
                run: config => {
                    const value = '<access key id>'
                    const edits = setProperty(config, ['accessKeyID'], value, defaultFormattingOptions)
                    return { edits, selectText: value }
                },
            },
            {
                id: 'setSecretAccessKey',
                label: 'Set secret access key',
                run: config => {
                    const value = '<secret access key>'
                    const edits = setProperty(config, ['secretAccessKey'], value, defaultFormattingOptions)
                    return { edits, selectText: value }
                },
            },
            {
                id: 'setRegion',
                label: 'Set region',
                run: config => {
                    const value = '<region>'
                    const edits = setProperty(config, ['region'], value, defaultFormattingOptions)
                    return { edits, selectText: value }
                },
            },
            {
                id: 'excludeRepo',
                label: 'Exclude a repository',
                run: config => {
                    const value = { name: '<owner>/<repository>' }
                    const edits = setProperty(config, ['exclude', -1], value, defaultFormattingOptions)
                    return { edits, selectText: '{"name": "<owner>/<repository>"}' }
                },
            },
        ],
    },
    [GQL.ExternalServiceKind.BITBUCKETCLOUD]: {
        title: 'Bitbucket.org repositories',
        icon: BitbucketIcon,
        shortDescription: 'Add Bitbucket Cloud repositories.',
        jsonSchema: bitbucketCloudSchemaJSON,
        defaultDisplayName: 'Bitbucket Cloud',
        defaultConfig: `// Use Ctrl+Space for completion, and hover over JSON properties for documentation.
// Bitbucket Cloud external service docs: https://docs.sourcegraph.com/admin/external_service/bitbucket_cloud#configuration
{
  "url": "https://bitbucket.org",

  // The username the app password belongs to
  "username": "<username>",

  // An app password (https://confluence.atlassian.com/bitbucket/app-passwords-828781300.html) with read scope over the repositories and teams to be added to Sourcegraph
  "appPassword": "<app password>",

  // teams: List of teams whose repositories should be selected
  // "teams": [
  //   "<team name>"
  // ],
}`,
        editorActions: [
            {
                id: 'setAppPassword',
                label: 'Set app password',
                run: config => {
                    const value = '<app password>'
                    const edits = setProperty(config, ['appPassword'], value, defaultFormattingOptions)
                    return { edits, selectText: value }
                },
            },
        ],
    },
    [GQL.ExternalServiceKind.BITBUCKETSERVER]: {
        title: 'Bitbucket Server repositories',
        icon: BitbucketIcon,
        shortDescription: 'Add Bitbucket Server repositories.',
        jsonSchema: bitbucketServerSchemaJSON,
        defaultDisplayName: 'Bitbucket Server',
        defaultConfig: `// Use Ctrl+Space for completion, and hover over JSON properties for documentation.
// Bitbucket Server external service docs: https://docs.sourcegraph.com/admin/external_service/bitbucket_server#configuration
{
  "url": "https://bitbucket.example.com",

  // Create a personal access token with read scope at
  // https://<bitbucket-hostname>/plugins/servlet/access-tokens/add
  "token": "<access token>",

  // The username the token belongs to
  "username": "<username>",

  // SELECTING REPOSITORIES
  //
  // There are 3 fields used to select repositories for searching and code intel:
  //  - repositoryQuery (required)
  //  - repos
  //  - exclude

  // repositoryQuery: List of strings: a special keyword "none" (which disables querying),
  // "all" (which selects all repositories visible to the given token), or any repository
  // search query parameters (e.g "?name=<repo name>&projectname=<project>&visibility=private")
  // See the list of parameters at: https://docs.atlassian.com/bitbucket-server/rest/6.1.2/bitbucket-rest.html#idp355
  "repositoryQuery": [],

  // repos: Explicit list of repositories to select
  // "repos": [
  //   "<project/<repository>"
  // ],

  // exclude: Repositories to exclude (overrides repositories from repositoryQuery and repos)
  // "exclude": [
  //   {
  //     "name": "<project/<repository>"
  //   }
  // ]
}`,
        editorActions: [
            {
                id: 'setURL',
                label: 'Set Bitbucket Server URL',
                run: config => {
                    const value = 'https://bitbucket.example.com'
                    const edits = setProperty(config, ['url'], value, defaultFormattingOptions)
                    return { edits, selectText: value }
                },
            },
            {
                id: 'setPersonalAccessToken',
                label: 'Set access token',
                run: config => {
                    const value = '<access token>'
                    const edits = setProperty(config, ['token'], value, defaultFormattingOptions)
                    return { edits, selectText: value }
                },
            },
            {
                id: 'addProjectRepos',
                label: 'Add project repositories',
                run: config => {
                    const value = '?projectname=<project>'
                    const edits = setProperty(config, ['repositoryQuery', -1], value, defaultFormattingOptions)
                    return { edits, selectText: '<project>' }
                },
            },
            {
                id: 'addRepo',
                label: 'Add a repository',
                run: config => {
                    const value = '<project/<repository>'
                    const edits = setProperty(config, ['repos', -1], value, defaultFormattingOptions)
                    return { edits, selectText: '<project/<repository>' }
                },
            },
            {
                id: 'excludeRepo',
                label: 'Exclude a repository',
                run: config => {
                    const value = { name: '<project/<repository>' }
                    const edits = setProperty(config, ['exclude', -1], value, defaultFormattingOptions)
                    return { edits, selectText: '{"name": "<project/<repository>"}' }
                },
            },
            {
                id: 'setSelfSignedCert',
                label: 'Set internal or self-signed certificate',
                run: config => {
                    const value = '<certificate>'
                    const edits = setProperty(config, ['certificate'], value, defaultFormattingOptions)
                    return { edits, selectText: value }
                },
            },
            {
                id: 'enablePermissions',
                label: 'Enforce permissions',
                run: config => {
                    const value = {
                        COMMENT_SENTINEL: true,
                        identityProvider: { type: 'username' },
                        oauth: {
                            consumerKey: '<consumer key>',
                            signingKey: '<signing key>',
                        },
                        ttl: '3h',
                        hardTTL: '72h',
                    }
                    const comment =
                        '// Follow setup instructions in https://docs.sourcegraph.com/admin/repo/permissions#bitbucket_server'
                    const edit = editWithComment(config, ['authorization'], value, comment)
                    return { edits: [edit], selectText: comment }
                },
            },
        ],
    },
    [GQL.ExternalServiceKind.GITLAB]: {
        title: 'GitLab projects',
        icon: GitLabIcon,
        shortDescription: 'Add GitLab projects.',
        jsonSchema: gitlabSchemaJSON,
        defaultDisplayName: 'GitLab',
        defaultConfig: `// Use Ctrl+Space for completion, and hover over JSON properties for documentation.
// GitLab external service docs: https://docs.sourcegraph.com/admin/external_service/gitlab#configuration
{
  "url": "https://example.gitlab.com",

  // Create a personal access token with api scope at https://[your-gitlab-hostname]/profile/personal_access_tokens
  "token": "<access token>",

  // SELECTING REPOSITORIES
  //
  // There are 3 fields used to select repositories for searching and code intel:
  //  - projectQuery (required)
  //  - projects
  //  - exclude

  // List of strings, either a special keyword "none" (which disables querying), search query parameters
  // such as "?search=sourcegraph", and "?visibility=private".
  //
  // For getting started, use the "Quick configure" buttons above the editor to build an initial set of
  // of queries.
  "projectQuery": [
  //   "?archived=no\u0026visibility=private" // set this to "none" to disable querying
  ],

  // projects: Project repositories to select.  Supports name: {"name": "group/name"}, or ID: {"id": 42})
  // "projects": [
  //   { "name": "<group>/<name>" },
  //   { "id": <id> }
  // ],

  // exclude: Project repositories to exclude.  Supports name: {"name": "group/name"}, or ID: {"id": 42})
  // "exclude": [
  //   { "name": "<group>/<name>" },
  //   { "id": <id> }
  // ]
}`,
        editorActions: [
            {
                id: 'setURL',
                label: 'Set GitLab URL',
                run: config => {
                    const value = 'https://gitlab.example.com'
                    const edits = setProperty(config, ['url'], value, defaultFormattingOptions)
                    return { edits, selectText: value }
                },
            },
            {
                id: 'setPersonalAccessToken',
                label: 'Set access token',
                run: config => {
                    const value = '<access token>'
                    const edits = setProperty(config, ['token'], value, defaultFormattingOptions)
                    return { edits, selectText: value }
                },
            },
            {
                id: 'setSelfSignedCert',
                label: 'Set internal or self-signed certificate',
                run: config => {
                    const value = '<certificate>'
                    const edits = setProperty(config, ['certificate'], value, defaultFormattingOptions)
                    return { edits, selectText: value }
                },
            },
            {
                id: 'syncInternalProjects',
                label: 'Sync internal projects',
                run: config => {
                    const value = '?visibility=internal'
                    const edits = setProperty(config, ['projectQuery', -1], value, defaultFormattingOptions)
                    return { edits, selectText: value }
                },
            },
            {
                id: 'syncPrivateProjects',
                label: 'Sync private projects',
                run: config => {
                    const value = '?visibility=private'
                    const edits = setProperty(config, ['projectQuery', -1], value, defaultFormattingOptions)
                    return { edits, selectText: value }
                },
            },
            {
                id: 'syncPublicProjects',
                label: 'Sync public projects',
                run: config => {
                    const value = '?visibility=public'
                    const edits = setProperty(config, ['projectQuery', -1], value, defaultFormattingOptions)
                    return { edits, selectText: value }
                },
            },
            {
                id: 'syncGroupProjects',
                label: 'Sync group projects',
                run: config => {
                    const value = 'groups/<group ID>/projects'
                    const edits = setProperty(config, ['projectQuery', -1], value, defaultFormattingOptions)
                    return { edits, selectText: '<group ID>' }
                },
            },
            {
                id: 'syncMembershipProjects',
                label: 'Sync all projects the access token user is a member of',
                run: config => {
                    const value = '?membership=true'
                    const edits = setProperty(config, ['projectQuery', -1], value, defaultFormattingOptions)
                    return { edits, selectText: value }
                },
            },
            {
                id: 'syncProjectsMatchingSearch',
                label: 'Sync projects matching search',
                run: config => ({
                    edits: setProperty(
                        config,
                        ['projectQuery', -1],
                        '?search=<search query>',
                        defaultFormattingOptions
                    ),
                    selectText: '<search query>',
                }),
            },
            {
                id: 'enforcePermissionsOAuth',
                label: 'Enforce permissions (OAuth)',
                run: config => {
                    const value = {
                        identityProvider: {
                            COMMENT_SENTINEL: true,
                            type: 'oauth',
                        },
                    }
                    const comment = editorActionComments.enforcePermissionsOAuth
                    const edit = editWithComment(config, ['authorization'], value, comment)
                    return { edits: [edit], selectText: comment }
                },
            },
            {
                id: 'enforcePermissionsSSO',
                label: 'Enforce permissions (SSO)',
                run: config => {
                    const value = {
                        COMMENT_SENTINEL: true,
                        identityProvider: {
                            type: 'external',
                            authProviderID: '<configID field of the auth provider>',
                            authProviderType: '<type field of the auth provider>',
                            gitlabProvider:
                                '<name that identifies the auth provider to GitLab (hover over "gitlabProvider" for docs)>',
                        },
                    }
                    const comment = editorActionComments.enforcePermissionsSSO
                    const edit = editWithComment(config, ['authorization'], value, comment)
                    return { edits: [edit], selectText: comment }
                },
            },
            {
                id: 'addProject',
                label: 'Add a project',
                run: config => {
                    const value = { name: '<group>/<project>' }
                    const edits = setProperty(config, ['projects', -1], value, defaultFormattingOptions)
                    return { edits, selectText: '{"name": "<group>/<project>"}' }
                },
            },
            {
                id: 'excludeProject',
                label: 'Exclude a project',
                run: config => {
                    const value = { name: '<group>/<project>' }
                    const edits = setProperty(config, ['exclude', -1], value, defaultFormattingOptions)
                    return { edits, selectText: '{"name": "<group>/<project>"}' }
                },
            },
        ],
    },
    [GQL.ExternalServiceKind.GITOLITE]: {
        title: 'Gitolite repositories',
        icon: GitIcon,
        shortDescription: 'Add Gitolite repositories.',
        jsonSchema: gitoliteSchemaJSON,
        defaultDisplayName: 'Gitolite',
        defaultConfig: `{
  // Use Ctrl+Space for completion, and hover over JSON properties for documentation.
  // Configuration options are documented here:
  // https://docs.sourcegraph.com/admin/external_service/gitolite#configuration

  "prefix": "gitolite.example.com/",
  "host": "git@gitolite.example.com"
}`,
        editorActions: [
            {
                id: 'setPrefix',
                label: 'Set prefix',
                run: config => {
                    const value = 'gitolite.example.com/'
                    const edits = setProperty(config, ['prefix'], value, defaultFormattingOptions)
                    return { edits, selectText: value }
                },
            },
            {
                id: 'setHost',
                label: 'Set host',
                run: config => {
                    const value = 'git@gitolite.example.com'
                    const edits = setProperty(config, ['host'], value, defaultFormattingOptions)
                    return { edits, selectText: value }
                },
            },
        ],
    },
    [GQL.ExternalServiceKind.PHABRICATOR]: {
        title: 'Phabricator connection',
        icon: PhabricatorIcon,
        shortDescription:
            'Associate Phabricator repositories with existing repositories on Sourcegraph. Mirroring is not supported.',
        jsonSchema: phabricatorSchemaJSON,
        defaultDisplayName: 'Phabricator',
        defaultConfig: `{
  // Use Ctrl+Space for completion, and hover over JSON properties for documentation.
  // Configuration options are documented here:
  // https://docs.sourcegraph.com/admin/external_service/phabricator#configuration

  "url": "https://phabricator.example.com",
  "token": "",
  "repos": []
}`,
        editorActions: [
            {
                id: 'setPhabricatorURL',
                label: 'Set Phabricator URL',
                run: config => {
                    const value = 'https://phabricator.example.com'
                    const edits = setProperty(config, ['url'], value, defaultFormattingOptions)
                    return { edits, selectText: value }
                },
            },
            {
                id: 'setAccessToken',
                label: 'Set Phabricator access token',
                run: config => {
                    const value = '<Phabricator access token>'
                    const edits = setProperty(config, ['token'], value, defaultFormattingOptions)
                    return { edits, selectText: value }
                },
            },
            {
                id: 'addRepository',
                label: 'Add a repository',
                run: config => {
                    const value = {
                        callsign: '<Phabricator repository callsign>',
                        path: '<Sourcegraph repository full name>',
                    }
                    const edits = setProperty(config, ['repos', -1], value, defaultFormattingOptions)
                    return { edits, selectText: '<Phabricator repository callsign>' }
                },
            },
        ],
    },
    [GQL.ExternalServiceKind.OTHER]: {
        title: 'Single Git repositories',
        icon: GitIcon,
        shortDescription: 'Add single Git repositories by clone URL.',
        jsonSchema: otherExternalServiceSchemaJSON,
        defaultDisplayName: 'Git repositories',
        defaultConfig: `{
  // Use Ctrl+Space for completion, and hover over JSON properties for documentation.
  // Configuration options are documented here:
  // https://docs.sourcegraph.com/admin/external_service/other#configuration

  // Supported URL schemes are: http, https, git and ssh
  "url": "https://my-other-githost.example.com",

  // Repository clone paths may be relative to the url (preferred) or absolute.
  "repos": []
}`,
        editorActions: [
            {
                id: 'setURL',
                label: 'Set Git host URL',
                run: config => {
                    const value = 'https://my-other-githost.example.com'
                    const edits = setProperty(config, ['url'], value, defaultFormattingOptions)
                    return { edits, selectText: value }
                },
            },
            {
                id: 'addRepo',
                label: 'Add a repository',
                run: config => {
                    const value = 'path/to/repository'
                    const edits = setProperty(config, ['repos', -1], value, defaultFormattingOptions)
                    return { edits, selectText: value }
                },
            },
        ],
    },
}

/**
 * Some external services have variants that should be presented in the UI in a different way
 * but are not fundamentally different from one another. This type defines the allowed variant
 * values.
 */
export type ExternalServiceVariant = 'dotcom' | 'enterprise'

export function isExternalServiceVariant(s: string): s is ExternalServiceVariant {
    return s === 'dotcom' || s === 'enterprise'
}

export interface AddExternalServiceMetadata extends ExternalServiceKindMetadata {
    kind: GQL.ExternalServiceKind
    variant?: ExternalServiceVariant
}

/**
 * We want to have more than one "add" option for some external services (e.g., GitHub.com vs. GitHub Enterprise).
 * These patches define the overrides that should be applied to certain external services.
 */
const externalServiceAddVariants: Partial<Record<
    GQL.ExternalServiceKind,
    Partial<Record<ExternalServiceVariant, Partial<ExternalServiceKindMetadata>>>
>> = {
    [GQL.ExternalServiceKind.GITHUB]: {
        dotcom: {
            title: 'GitHub.com repositories',
            shortDescription: 'Add GitHub.com repositories.',
            editorActions: GITHUB_EXTERNAL_SERVICE.editorActions || [],
        },
        enterprise: {
            title: 'GitHub Enterprise repositories',
            shortDescription: 'Add GitHub Enterprise repositories.',
            defaultDisplayName: 'GitHub Enterprise',
            defaultConfig: `// Use Ctrl+Space for completion, and hover over JSON properties for documentation.
// GitHub external service docs: https://docs.sourcegraph.com/admin/external_service/github
{
  // GitHub Enterprise URL
  "url": "https://github.example.com",

  // token: GitHub API access token.
  // Visit https://[github-enterprise-url]/settings/tokens/new?scopes=repo&description=Sourcegraph to create a token
  // with access to public and private repositories
  "token": "<access token>",

  // SELECTING REPOSITORIES
  //
  // There are 3 fields used to select repositories for searching and code intel:
  //  - repositoryQuery (required)
  //  - repos
  //  - exclude
  //

  // repositoryQuery: List of strings, either a special keyword, e.g. "affiliated", or
  // GitHub search qualifiers, e.g. "archived:false"
  //
  // For getting started, use either:
  //  - "org:<name>" // (e.g. "org:sourcegraph") all repositories belonging to the organization
  // or
  //  - "affiliated" // all repositories affiliated (accessible) by the token's owner
  //
  // Additional query strings can be added to refine results:
  //  - "archived:false fork:no created:>=2016" // use of multiple search qualifiers
  //  - "user:docker repo:kubernetes/kubernetes" // fetch repositories outside of the user/org account
  //
  // See https://help.github.com/en/articles/searching-for-repositories for the list of search qualifiers.
  "repositoryQuery": [
  //   "org:name"
  ],

  // repos: Explicit list of repositories to select
  // "repos": [
  //   "<owner>/<repository>"
  // ],

  // exclude: Repositories to exclude (overrides repositories from repositoryQuery and repos)
  // "exclude": [
  //   {
  //   "name": "<owner>/<repository>"
  //   }
  // ]
}`,
        },
    },
}

export const ALL_EXTERNAL_SERVICE_ADD_VARIANTS: AddExternalServiceMetadata[] = flatMap(
    map(ALL_EXTERNAL_SERVICES, (service: ExternalServiceKindMetadata, kindString: string):
        | AddExternalServiceMetadata
        | AddExternalServiceMetadata[] => {
        const kind = kindString as GQL.ExternalServiceKind
        if (externalServiceAddVariants[kind]) {
            const patches = externalServiceAddVariants[kind]
            return map(patches, (patch, variantString) => {
                const variant = variantString as ExternalServiceVariant
                return {
                    ...service,
                    kind,
                    variant,
                    ...patch,
                }
            })
        }
        return {
            ...service,
            kind,
        }
    })
)

export function getExternalService(
    kind: GQL.ExternalServiceKind,
    variantForAdd?: ExternalServiceVariant
): ExternalServiceKindMetadata {
    const foundVariants = ALL_EXTERNAL_SERVICE_ADD_VARIANTS.filter(
        serviceVariant => serviceVariant.kind === kind && serviceVariant.variant === variantForAdd
    )
    if (foundVariants.length > 0) {
        return foundVariants[0]
    }
    return ALL_EXTERNAL_SERVICES[kind]
}
