import { FormattingOptions } from '@sqs/jsonc-parser'
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

export const GITHUB_EXTERNAL_SERVICE: ExternalServiceKindMetadata = {
    title: 'GitHub repositories',
    icon: <GithubCircleIcon size={ICON_SIZE} />,
    jsonSchema: githubSchemaJSON,
    editorActions: githubEditorActions,
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
  // Configuration options are documented here:
  // https://docs.sourcegraph.com/admin/external_service/github#configuration

  "url": "https://github.com",

  // A token is required for access to private repos, but is also helpful for public repos
  // because it grants a higher hourly rate limit to Sourcegraph.
  // Create one with the repo scope at https://[your-github-instance]/settings/tokens/new
  "token": "",

  // An array of strings specifying which GitHub or GitHub Enterprise repositories to mirror on Sourcegraph.
  // See https://docs.sourcegraph.com/admin/site_config/all#githubconnection-object for more details.
  "repositoryQuery": [
      "none"
  ],

  // Sync public repositories from https://github.com by adding them to "repos".
  // (This is not necessary for GitHub Enterprise instances)
  "repos": [
      // "sourcegraph/sourcegraph"
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

  // Create a personal access token with read scope at
  // https://[your-bitbucket-hostname]/plugins/servlet/access-tokens/add
  "token": ""
}`,
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
  "token": ""
}`,
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
    },
    [GQL.ExternalServiceKind.PHABRICATOR]: {
        title: 'Phabricator connection',
        icon: <PhabricatorIcon size={ICON_SIZE} />,
        iconBrandColor: 'phabricator',
        shortDescription: 'Add links to Phabricator from Sourcegraph.',
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
