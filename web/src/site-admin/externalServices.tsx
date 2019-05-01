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
    icon: JSX.Element | string

    /**
     * Color to display next to the icon in the external service "button"
     */
    iconBrandColor: 'github' | 'aws' | 'bitbucket' | 'gitlab' | 'gitolite' | 'phabricator' | 'git'

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

const ICON_SIZE = 45

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
        '// Prerequisite: you must configure GitHub as an OAuth auth provider in the critical site config (https://docs.sourcegraph.com/admin/auth#github). Otherwise, access to all repositories will be disallowed.',
    enforcePermissionsOAuth: `// Prerequisite: you must first update the critical site configuration to
    // include GitLab OAuth as an auth provider.
    // See https://docs.sourcegraph.com/admin/auth#gitlab for instructions.`,
    enforcePermissionsSSO: `// Prerequisite: You will need a sudo-level access token. If you can configure
    // GitLab as an OAuth identity provider for Sourcegraph, we recommend that
    // option instead.
    //
    // 1. Ensure the personal access token in this config has admin privileges
    //    (https://docs.gitlab.com/ee/api/#sudo).
    // 2. Update the critical site configuration in the management console to
    //    include the SSO auth provider for GitLab (https://docs.sourcegraph.com/admin/auth).
    // 3. Update the fields below to match the properties of this auth provider
    //    (https://docs.sourcegraph.com/admin/repo/permissions#sudo-access-token).`,
}

export const GITHUB_EXTERNAL_SERVICE: ExternalServiceKindMetadata = {
    title: 'GitHub repositories',
    icon: <GithubCircleIcon size={ICON_SIZE} />,
    jsonSchema: githubSchemaJSON,
    editorActions: [
        {
            id: 'setAccessToken',
            label: 'Set access token',
            run: config => {
                const value = '<GitHub personal access token>'
                const edits = setProperty(config, ['token'], value, defaultFormattingOptions)
                return { edits, selectText: '<GitHub personal access token>' }
            },
        },
        {
            id: 'addOrgRepo',
            label: 'Add organization repositories',
            run: config => {
                const value = 'org:<organization name>'
                const edits = setProperty(config, ['repositoryQuery', -1], value, defaultFormattingOptions)
                return { edits, selectText: '<organization name>' }
            },
        },
        {
            id: 'addRepo',
            label: 'Add a repository',
            run: config => {
                const value = '<GitHub owner>/<GitHub repository name>'
                const edits = setProperty(config, ['repos', -1], value, defaultFormattingOptions)
                return { edits, selectText: '<GitHub owner>/<GitHub repository name>' }
            },
        },
        {
            id: 'excludeRepo',
            label: 'Exclude a repository',
            run: config => {
                const value = { name: '<GitHub owner>/<GitHub repository name>' }
                const edits = setProperty(config, ['exclude', -1], value, defaultFormattingOptions)
                return { edits, selectText: '{"name": "<GitHub owner>/<GitHub repository name>"}' }
            },
        },
        {
            id: 'addSearchQueryRepos',
            label: 'Add repositories matching search query',
            run: config => {
                const value = '<GitHub search query>'
                const edits = setProperty(config, ['repositoryQuery', -1], value, defaultFormattingOptions)
                return { edits, selectText: '<GitHub search query>' }
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
    iconBrandColor: 'github',
    shortDescription: 'Add GitHub repositories.',
    longDescription: (
        <span>
            Adding this configuration enables Sourcegraph to sync repositories from GitHub. Click the "quick configure"
            buttons for common actions or directly edit the JSON configuration.{' '}
            <Link target="_blank" to="/help/admin/external_service/github#configuration">
                Read the docs
            </Link>{' '}
            for more info about each field.
        </span>
    ),
    defaultDisplayName: 'GitHub',
    defaultConfig: `{
  // Use Ctrl+Space for completion, and hover over JSON properties for documentation.
  // Docs: https://docs.sourcegraph.com/admin/external_service/github#configuration

  "url": "https://github.com", // change to use with GitHub Enterprise

  // Enter an access token to mirror GitHub repositories. Create one for GitHub.com at
  // https://github.com/settings/tokens/new?scopes=repo&description=Sourcegraph
  // (for GitHub Enterprise, replace github.com with your instance's hostname).
  // The "repo" scope is required to mirror private repositories.
  "token": "",

  // An array of strings specifying which GitHub or GitHub Enterprise repositories to mirror on Sourcegraph.
  // See the repositoryQuery documentation at https://docs.sourcegraph.com/admin/external_service/github#configuration for details.
  "repositoryQuery": [
    // "org:sourcegraph"
  ]
}`,
}

export const ALL_EXTERNAL_SERVICES: Record<GQL.ExternalServiceKind, ExternalServiceKindMetadata> = {
    [GQL.ExternalServiceKind.GITHUB]: GITHUB_EXTERNAL_SERVICE,
    [GQL.ExternalServiceKind.AWSCODECOMMIT]: {
        title: 'AWS CodeCommit repositories',
        icon: <AmazonIcon size={ICON_SIZE} />,
        iconBrandColor: 'aws',
        shortDescription: 'Add AWS CodeCommit repositories.',
        jsonSchema: awsCodeCommitSchemaJSON,
        defaultDisplayName: 'AWS CodeCommit',
        defaultConfig: `{
  // Use Ctrl+Space for completion, and hover over JSON properties for documentation.
  // Configuration options are documented here:
  // https://docs.sourcegraph.com/admin/external_service/aws_codecommit#configuration

  "region": "",
  "accessKeyID": "",
  "secretAccessKey": ""
}`,
        editorActions: [
            {
                id: 'setRegion',
                label: 'Set AWS region',
                run: config => {
                    const value = '<AWS region>'
                    const edits = setProperty(config, ['region'], value, defaultFormattingOptions)
                    return { edits, selectText: value }
                },
            },
            {
                id: 'setAccessKeyID',
                label: 'Set access key ID',
                run: config => {
                    const value = '<AWS access key ID>'
                    const edits = setProperty(config, ['accessKeyID'], value, defaultFormattingOptions)
                    return { edits, selectText: value }
                },
            },
            {
                id: 'setSecretAccessKey',
                label: 'Set AWS secret access key',
                run: config => {
                    const value = '<AWS secret access key>'
                    const edits = setProperty(config, ['secretAccessKey'], value, defaultFormattingOptions)
                    return { edits, selectText: value }
                },
            },
        ],
    },
    [GQL.ExternalServiceKind.BITBUCKETSERVER]: {
        title: 'Bitbucket Server repositories',
        icon: <BitbucketIcon size={ICON_SIZE} />,
        iconBrandColor: 'bitbucket',
        shortDescription: 'Add Bitbucket Server repositories.',
        jsonSchema: bitbucketServerSchemaJSON,
        defaultDisplayName: 'Bitbucket Server',
        defaultConfig: `{
  // Use Ctrl+Space for completion, and hover over JSON properties for documentation.
  // Configuration options are documented here:
  // https://docs.sourcegraph.com/admin/external_service/bitbucket_server#configuration

  "url": "https://bitbucket.example.com",

  // The username of the user that owns the token defined below.
  "username": "",

  // Create a personal access token with read scope at
  // https://[your-bitbucket-hostname]/plugins/servlet/access-tokens/add
  "token": "",

  // An array of strings specifying which repositories to mirror on Sourcegraph.
  // Each string is a URL query string with parameters that filter the list of returned repos.
  // Example: "?name=my-repo&projectname=PROJECT&visibility=private".
  //
  // The special string "none" can be used as the only element to disable this feature.
  // Repositories matched by multiple query strings are only imported once.
  //
  // Here's the official Bitbucket Server documentation about which query string parameters are valid:
  // https://docs.atlassian.com/bitbucket-server/rest/6.1.2/bitbucket-rest.html#idp355
  "repositoryQuery": [
    // "?name=sourcegraph"
  ]
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
                label: 'Set personal access token',
                run: config => {
                    const value = '<Bitbucket Server personal access token>'
                    const edits = setProperty(config, ['token'], value, defaultFormattingOptions)
                    return { edits, selectText: value }
                },
            },
            {
                id: 'setSelfSignedCert',
                label: 'Set internal or self-signed certificate',
                run: config => {
                    const value = '<internal-CA- or self-signed certificate>'
                    const edits = setProperty(config, ['certificate'], value, defaultFormattingOptions)
                    return { edits, selectText: value }
                },
            },
            {
                id: 'addProjectRepos',
                label: 'Add project repositories',
                run: config => {
                    const value = '?projectname=<project name>'
                    const edits = setProperty(config, ['repositoryQuery', -1], value, defaultFormattingOptions)
                    return { edits, selectText: '<project name>' }
                },
            },
            {
                id: 'addRepo',
                label: 'Add a repository',
                run: config => {
                    const value = '<projectKey>/<repoSlug>'
                    const edits = setProperty(config, ['repos', -1], value, defaultFormattingOptions)
                    return { edits, selectText: '<projectKey>/<repoSlug>' }
                },
            },
            {
                id: 'excludeRepo',
                label: 'Exclude a repository',
                run: config => {
                    const value = { name: '<projectKey>/<repoSlug>' }
                    const edits = setProperty(config, ['exclude', -1], value, defaultFormattingOptions)
                    return { edits, selectText: '{"name": "<projectKey>/<repoSlug>"}' }
                },
            },
        ],
    },
    [GQL.ExternalServiceKind.GITLAB]: {
        title: 'GitLab projects',
        icon: <GitLabIcon size={ICON_SIZE} />,
        iconBrandColor: 'gitlab',
        shortDescription: 'Add GitLab projects.',
        jsonSchema: gitlabSchemaJSON,
        defaultDisplayName: 'GitLab',
        defaultConfig: `{
  // Use Ctrl+Space for completion, and hover over JSON properties for documentation.
  // Configuration options are documented here:
  // https://docs.sourcegraph.com/admin/external_service/gitlab#configuration

  "url": "https://gitlab.example.com",

  // Create a personal access token with api scope at
  // https://[your-gitlab-hostname]/profile/personal_access_tokens
  "token": "",

  // An array of strings specifying GitLab project search queries to mirror on Sourcegraph.
  // See the projectQuery documentation at https://docs.sourcegraph.com/admin/external_service/gitlab#configuration for details.
  "projectQuery": [
    // "?search=sourcegraph",
  ]
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
                label: 'Set personal access token',
                run: config => {
                    const value = '<GitLab personal access token>'
                    const edits = setProperty(config, ['token'], value, defaultFormattingOptions)
                    return { edits, selectText: value }
                },
            },
            {
                id: 'setSelfSignedCert',
                label: 'Set internal or self-signed certificate',
                run: config => {
                    const value = '<internal-CA- or self-signed certificate>'
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
                    const value = { name: '<GitLab Group>/<Project Name>' }
                    const edits = setProperty(config, ['projects', -1], value, defaultFormattingOptions)
                    return { edits, selectText: '{"name": "<GitLab Group>/<Project Name>"}' }
                },
            },
            {
                id: 'excludeProject',
                label: 'Exclude a project',
                run: config => {
                    const value = { name: '<GitLab Group>/<Project Name>' }
                    const edits = setProperty(config, ['exclude', -1], value, defaultFormattingOptions)
                    return { edits, selectText: '{"name": "<GitLab Group>/<Project Name>"}' }
                },
            },
        ],
    },
    [GQL.ExternalServiceKind.GITOLITE]: {
        title: 'Gitolite repositories',
        icon: <GitIcon size={ICON_SIZE} />,
        iconBrandColor: 'gitolite',
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
        icon: <PhabricatorIcon size={ICON_SIZE} />,
        iconBrandColor: 'phabricator',
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
        icon: <GitIcon size={ICON_SIZE} />,
        iconBrandColor: 'git',
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
    serviceKind: GQL.ExternalServiceKind
    variant?: ExternalServiceVariant
}

/**
 * We want to have more than one "add" option for some external services (e.g., GitHub.com vs. GitHub Enterprise).
 * These patches define the overrides that should be applied to certain external services.
 */
const externalServiceAddVariants: Partial<
    Record<GQL.ExternalServiceKind, Partial<Record<ExternalServiceVariant, Partial<ExternalServiceKindMetadata>>>>
> = {
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
            defaultConfig: `{
  // Use Ctrl+Space for completion, and hover over JSON properties for documentation.
  // Configuration options are documented here:
  // https://docs.sourcegraph.com/admin/external_service/github#configuration

  // Set this to the URL for your GitHub Enterprise.
  "url": "https://github.example.com",

  // A token is required for access to private repos, but is also helpful for public repos
  // because it grants a higher hourly rate limit to Sourcegraph.
  // Create one with the repo scope at https://[your-github-instance]/settings/tokens/new
  "token": ""
}`,
        },
    },
}

export const ALL_EXTERNAL_SERVICE_ADD_VARIANTS: AddExternalServiceMetadata[] = flatMap(
    map(
        ALL_EXTERNAL_SERVICES,
        (
            service: ExternalServiceKindMetadata,
            kindString: string
        ): AddExternalServiceMetadata | AddExternalServiceMetadata[] => {
            const kind = kindString as GQL.ExternalServiceKind
            if (externalServiceAddVariants[kind]) {
                const patches = externalServiceAddVariants[kind]
                return map(patches, (patch, variantString) => {
                    const variant = variantString as ExternalServiceVariant
                    return {
                        ...service,
                        serviceKind: kind,
                        variant,
                        ...patch,
                    }
                })
            }
            return {
                ...service,
                serviceKind: kind,
            }
        }
    )
)

export function getExternalService(
    kind: GQL.ExternalServiceKind,
    variantForAdd?: ExternalServiceVariant
): ExternalServiceKindMetadata {
    const foundVariants = ALL_EXTERNAL_SERVICE_ADD_VARIANTS.filter(
        serviceVariant => serviceVariant.serviceKind === kind && serviceVariant.variant === variantForAdd
    )
    if (foundVariants.length > 0) {
        return foundVariants[0]
    }
    return ALL_EXTERNAL_SERVICES[kind]
}
