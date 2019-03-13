import { FormattingOptions } from '@sqs/jsonc-parser'
import { setProperty } from '@sqs/jsonc-parser/lib/edit'
import { flatMap, map } from 'lodash'
import AmazonIcon from 'mdi-react/AmazonIcon'
import BitbucketIcon from 'mdi-react/BitbucketIcon'
import GithubCircleIcon from 'mdi-react/GithubCircleIcon'
import GitIcon from 'mdi-react/GitIcon'
import GitLabIcon from 'mdi-react/GitlabIcon'
import React from 'react'
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
export interface ExternalServiceCategory {
    /**
     * Title to show in the external service "button"
     */
    title: string

    /**
     * Icon to show in the external service "button"
     */
    icon: JSX.Element | string

    /**
     * Color to display in the external service "button"
     */
    color: string

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

const iconSize = 50

const githubEditorActions: EditorAction[] = [
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
        id: 'addSingleRepo',
        label: 'Add single repository',
        run: config => {
            const value = '<GitHub owner>/<GitHub repository name>'
            const edits = setProperty(config, ['repos', -1], value, defaultFormattingOptions)
            return { edits, selectText: '<GitHub owner>/<GitHub repository name>' }
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
]

export const GITHUB_EXTERNAL_SERVICE: ExternalServiceCategory = {
    title: 'GitHub repositories',
    icon: <GithubCircleIcon size={iconSize} />,
    jsonSchema: githubSchemaJSON,
    editorActions: githubEditorActions,
    color: '#2ebc4f',
    shortDescription: 'Add GitHub repositories.',
    longDescription: (
        <span>
            Adding this configuration enables Sourcegraph to sync repositories from GitHub. Click the "quick configure"
            buttons for common actions or directly edit the JSON configuration.{' '}
            <a
                target="_blank"
                href="https://docs.sourcegraph.com/integration/github#github-integration-with-sourcegraph"
            >
                Read the docs
            </a>{' '}
            for more info about each field.
        </span>
    ),
    defaultDisplayName: 'GitHub',
    defaultConfig: `{
  // Use Ctrl+Space for completion, and hover over JSON properties for documentation.
  // Configuration options are documented here:
  // https://docs.sourcegraph.com/admin/site_config/all#githubconnection-object

  "url": "https://github.com",

  // A token is required for access to private repos, but is also helpful for public repos
  // because it grants a higher hourly rate limit to Sourcegraph.
  // Create one with the repo scope at https://[your-github-instance]/settings/tokens/new
  "token": "",

  // Sync public repositories from https://github.com by adding them to "repos".
  // (This is not necessary for GitHub Enterprise instances)
  // "repos": [
  //     "sourcegraph/sourcegraph"
  // ]

}`,
}

export const ALL_EXTERNAL_SERVICES: Record<GQL.ExternalServiceKind, ExternalServiceCategory> = {
    [GQL.ExternalServiceKind.GITHUB]: GITHUB_EXTERNAL_SERVICE,
    [GQL.ExternalServiceKind.AWSCODECOMMIT]: {
        title: 'AWS CodeCommit repositories',
        icon: <AmazonIcon size={iconSize} />,
        color: '#f8991d',
        shortDescription: 'Add AWS CodeCommit repositories.',
        jsonSchema: awsCodeCommitSchemaJSON,
        defaultDisplayName: 'AWS CodeCommit',
        defaultConfig: `{
  // Use Ctrl+Space for completion, and hover over JSON properties for documentation.
  // Configuration options are documented here:
  // https://docs.sourcegraph.com/admin/site_config/all#awscodecommitconnection-object

  "region": "",
  "accessKeyID": "",
  "secretAccessKey": ""
}`,
    },
    [GQL.ExternalServiceKind.BITBUCKETSERVER]: {
        title: 'Bitbucket Server repositories',
        icon: <BitbucketIcon size={iconSize} />,
        color: '#2684ff',
        shortDescription: 'Add Bitbucket Server repositories.',
        jsonSchema: bitbucketServerSchemaJSON,
        defaultDisplayName: 'Bitbucket Server',
        defaultConfig: `{
  // Use Ctrl+Space for completion, and hover over JSON properties for documentation.
  // Configuration options are documented here:
  // https://docs.sourcegraph.com/admin/site_config/all#bitbucketserverconnection-object

  "url": "https://bitbucket.example.com",

  // Create a personal access token with read scope at
  // https://[your-bitbucket-hostname]/plugins/servlet/access-tokens/add
  "token": ""
}`,
    },
    [GQL.ExternalServiceKind.GITLAB]: {
        title: 'GitLab projects',
        icon: <GitLabIcon size={iconSize} />,
        color: '#fc6e26',
        shortDescription: 'Add GitLab projects.',
        jsonSchema: gitlabSchemaJSON,
        defaultDisplayName: 'GitLab',
        defaultConfig: `{
  // Use Ctrl+Space for completion, and hover over JSON properties for documentation.
  // Configuration options are documented here:
  // https://docs.sourcegraph.com/admin/site_config/all#gitlabconnection-object

  "url": "https://gitlab.example.com",

  // Create a personal access token with api scope at
  // https://[your-gitlab-hostname]/profile/personal_access_tokens
  "token": ""
}`,
    },
    [GQL.ExternalServiceKind.GITOLITE]: {
        title: 'Gitolite repositories',
        icon: <GitIcon size={iconSize} />,
        color: '#e0e0e0',
        shortDescription: 'Add Gitolite repositories.',
        jsonSchema: gitoliteSchemaJSON,
        defaultDisplayName: 'Gitolite',
        defaultConfig: `{
  // Use Ctrl+Space for completion, and hover over JSON properties for documentation.
  // Configuration options are documented here:
  // https://docs.sourcegraph.com/admin/site_config/all#gitoliteconnection-object

  "prefix": "gitolite.example.com/",
  "host": "git@gitolite.example.com"
}`,
    },
    [GQL.ExternalServiceKind.PHABRICATOR]: {
        title: 'Phabricator connection',
        icon: <PhabricatorIcon size={iconSize} />,
        color: '#4a5f88',
        shortDescription: 'Add links to Phabricator from Sourcegraph.',
        jsonSchema: phabricatorSchemaJSON,
        defaultDisplayName: 'Phabricator',
        defaultConfig: `{
  // Use Ctrl+Space for completion, and hover over JSON properties for documentation.
  // Configuration options are documented here:
  // https://docs.sourcegraph.com/admin/site_config/all#phabricatorconnection-object

  "url": "https://phabricator.example.com",
  "token": "",
  "repos": []
}`,
    },
    [GQL.ExternalServiceKind.OTHER]: {
        title: 'Single Git repositories',
        icon: <GitIcon size={iconSize} />,
        color: '#f14e32',
        shortDescription: 'Add single Git repositories by clone URL.',
        jsonSchema: otherExternalServiceSchemaJSON,
        defaultDisplayName: 'Git repositories',
        defaultConfig: `{
  // Use Ctrl+Space for completion, and hover over JSON properties for documentation.
  // Configuration options are documented here:
  // https://docs.sourcegraph.com/admin/site_config/all#otherexternalserviceconnection-object

  // Supported URL schemes are: http, https, git and ssh
  "url": "https://my-other-githost.example.com",

  // Repository clone paths may be relative to the url (preferred) or absolute.
  "repos": []
}`,
    },
}

export function getExternalService(kind: GQL.ExternalServiceKind): ExternalServiceCategory {
    return ALL_EXTERNAL_SERVICES[kind]
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

export interface AddExternalServiceMetadata extends ExternalServiceCategory {
    serviceKind: GQL.ExternalServiceKind
    variant?: ExternalServiceVariant
}

/**
 * We want to have more than one "add" option for some external services (e.g., GitHub.com vs. GitHub Enterprise).
 * These patches define the overrides that should be applied to certain external services.
 */
const addPatches: Partial<
    Record<GQL.ExternalServiceKind, Partial<Record<ExternalServiceVariant, Partial<ExternalServiceCategory>>>>
> = {
    [GQL.ExternalServiceKind.GITHUB]: {
        dotcom: {
            title: 'GitHub.com repositories',
            shortDescription: 'Add GitHub.com repositories.',
            editorActions: [
                ...githubEditorActions,
                {
                    id: 'addPublicRepo',
                    label: 'Add public repository',
                    run: config => {
                        const value = '<GitHub owner>/<GitHub repository name>'
                        const edits = setProperty(config, ['repos', -1], value, defaultFormattingOptions)
                        return { edits, selectText: '<GitHub owner>/<GitHub repository name>' }
                    },
                },
            ],
        },
        enterprise: {
            title: 'GitHub Enterprise repositories',
            shortDescription: 'Add GitHub Enterprise repositories.',
        },
    },
}

export const ALL_ADD_EXTERNAL_SERVICES: AddExternalServiceMetadata[] = flatMap(
    map(
        ALL_EXTERNAL_SERVICES,
        (
            service: ExternalServiceCategory,
            kindString: string
        ): AddExternalServiceMetadata | AddExternalServiceMetadata[] => {
            const kind = kindString as GQL.ExternalServiceKind
            if (addPatches[kind]) {
                const patches = addPatches[kind]
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

const defaultFormattingOptions: FormattingOptions = {
    eol: '\n',
    insertSpaces: true,
    tabSize: 2,
}
