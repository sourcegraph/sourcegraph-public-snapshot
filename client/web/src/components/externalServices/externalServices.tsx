import React, { type ReactNode } from 'react'

import { type Edit, type JSONPath, type ModificationOptions, modify, parse as parseJSONC } from 'jsonc-parser'
import AwsIcon from 'mdi-react/AwsIcon'
import BitbucketIcon from 'mdi-react/BitbucketIcon'
import GithubIcon from 'mdi-react/GithubIcon'
import GitIcon from 'mdi-react/GitIcon'
import GitLabIcon from 'mdi-react/GitlabIcon'
import LanguageGoIcon from 'mdi-react/LanguageGoIcon'
import LanguageJavaIcon from 'mdi-react/LanguageJavaIcon'
import LanguagePythonIcon from 'mdi-react/LanguagePythonIcon'
import LanguageRubyIcon from 'mdi-react/LanguageRubyIcon'
import LanguageRustIcon from 'mdi-react/LanguageRustIcon'
import AzureDevOpsIcon from 'mdi-react/MicrosoftAzureDevopsIcon'
import NpmIcon from 'mdi-react/NpmIcon'

import { PerforceIcon, PhabricatorIcon } from '@sourcegraph/shared/src/components/icons'
import { Link, Code, Text } from '@sourcegraph/wildcard'

import awsCodeCommitSchemaJSON from '../../../../../schema/aws_codecommit.schema.json'
import azureDevOpsSchemaJSON from '../../../../../schema/azuredevops.schema.json'
import bitbucketCloudSchemaJSON from '../../../../../schema/bitbucket_cloud.schema.json'
import bitbucketServerSchemaJSON from '../../../../../schema/bitbucket_server.schema.json'
import gerritSchemaJSON from '../../../../../schema/gerrit.schema.json'
import githubSchemaJSON from '../../../../../schema/github.schema.json'
import gitlabSchemaJSON from '../../../../../schema/gitlab.schema.json'
import gitoliteSchemaJSON from '../../../../../schema/gitolite.schema.json'
import goModulesSchemaJSON from '../../../../../schema/go-modules.schema.json'
import jvmPackagesSchemaJSON from '../../../../../schema/jvm-packages.schema.json'
import npmPackagesSchemaJSON from '../../../../../schema/npm-packages.schema.json'
import otherExternalServiceSchemaJSON from '../../../../../schema/other_external_service.schema.json'
import pagureSchemaJSON from '../../../../../schema/pagure.schema.json'
import perforceSchemaJSON from '../../../../../schema/perforce.schema.json'
import phabricatorSchemaJSON from '../../../../../schema/phabricator.schema.json'
import pythonPackagesJSON from '../../../../../schema/python-packages.schema.json'
import rubyPackagesSchemaJSON from '../../../../../schema/ruby-packages.schema.json'
import rustPackagesJSON from '../../../../../schema/rust-packages.schema.json'
import {
    type ExternalRepositoryFields,
    ExternalServiceKind,
    ExternalServiceSyncJobState,
    type GitHubAppByAppIDResult,
} from '../../graphql-operations'
import type { EditorAction } from '../../settings/EditorActionsGroup'
import { GitHubAppSelector } from '../gitHubApps/GitHubAppSelector'

import type { ExternalServiceFieldsWithConfig } from './backend'
import { GerritIcon } from './GerritIcon'

/**
 * Metadata associated with adding a given external service.
 */
export interface AddExternalServiceOptions {
    kind: ExternalServiceKind

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
    shortDescription?: string

    /**
     * Instructions that will appear on the add / edit page
     */
    Instructions?: React.FunctionComponent

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

    /**
     * If present, denotes that we should show a status label, e.g. Beta or Experimental
     */
    status?: 'experimental' | 'beta'

    /**
     * If present, denotes that we should show additional fields on the form above the JSON
     */
    additionalFormComponent?: ReactNode
}

const defaultModificationOptions: ModificationOptions = {
    formattingOptions: {
        eol: '\n',
        insertSpaces: true,
        tabSize: 2,
    },
}

/**
 * editWithComment returns a Monaco edit action that sets the value of a JSON field with a
 * "//" comment annotating the field. The comment is inserted wherever
 * `"COMMENT_SENTINEL": true` appears in the JSON.
 */
function editWithComment(config: string, path: JSONPath, value: any, comment: string): Edit {
    const edit = modify(config, path, value, defaultModificationOptions)[0]
    edit.content = edit.content.replace('"COMMENT_SENTINEL": true', comment)
    return edit
}

const editorActionComments = {
    enablePermissions:
        '// Prerequisite: you must configure GitHub as an OAuth auth provider in the site config (https://sourcegraph.com/docs/admin/auth#github). Otherwise, access to all repositories will be disallowed.',
    enforcePermissionsOAuth: `// Prerequisite: you must first update the site configuration to
      // include GitLab OAuth as an auth provider.
      // See https://sourcegraph.com/docs/admin/auth#gitlab for instructions.`,
    enforcePermissionsSSO: `// Prerequisite: You need a sudo-level access token. If you can configure
    // GitLab as an OAuth identity provider for Sourcegraph, we recommend that
    // option instead.
    //
    // 1. Ensure the personal access token in this config has admin privileges
    //    (https://docs.gitlab.com/ee/api/#sudo).
    // 2. Update the site configuration to include the SSO auth provider for GitLab
    //    (https://sourcegraph.com/docs/admin/auth).
    // 3. Update the fields below to match the properties of this auth provider
    //    (https://sourcegraph.com/docs/admin/permissions).`,
}

const Field: React.FunctionComponent<{ children: React.ReactNode | string | string[] }> = props => (
    <Code className="hljs-type">{props.children}</Code>
)

const Value: React.FunctionComponent<{ children: React.ReactNode | string | string[] }> = props => (
    <Code className="hljs-attr">{props.children}</Code>
)

const GitHubInstructions: React.FunctionComponent<{ isEnterprise: boolean }> = ({ isEnterprise }) => (
    <div>
        <ol>
            {isEnterprise && (
                <li>
                    Set <Field>url</Field> to the URL of GitHub Enterprise.
                </li>
            )}
            <li>
                Create a GitHub access token (
                <Link
                    to="https://help.github.com/en/github/authenticating-to-github/creating-a-personal-access-token-for-the-command-line"
                    target="_blank"
                    rel="noopener noreferrer"
                >
                    instructions
                </Link>
                ) with <b>repo</b> scope.
                <li>
                    Set the value of the <Field>token</Field> field as your access token, in the configuration below.
                </li>
            </li>
            <li>
                Specify which repositories Sourcegraph should index using one of the following fields:
                <ul>
                    <li>
                        <Field>orgs</Field>: a list of GitHub organizations.
                    </li>
                    <li>
                        <Field>repositoryQuery</Field>: a list of GitHub search queries.
                        <br />
                        For example,
                        <Value>"org:sourcegraph created:&gt;2019-11-01"</Value> selects all repositories in organization
                        "sourcegraph" created after November 1, 2019.
                        <br />
                        You may also use <Value>"affiliated"</Value> to select all repositories affiliated with the
                        access token.
                    </li>
                    <li>
                        <Field>repos</Field>: a list of individual repositories.
                    </li>
                </ul>
            </li>
        </ol>
        <Text>
            See{' '}
            <Link rel="noopener noreferrer" target="_blank" to="/help/admin/code_hosts/github#configuration">
                the docs for more options
            </Link>
            , or try one of the buttons below.
        </Text>
    </div>
)

const GitHubAppInstructions: React.FunctionComponent = () => (
    <div>
        <ol>
            <li>
                Choose a GitHub App to populate the initial JSON configuration in <Code>gitHubAppDetails</Code>.
            </li>
            <li>
                If applicable, choose an associated installation. If your GitHub App has only one installation, this
                will be pre-populated along with other GitHub App information in <Code>orgs</Code>.
            </li>
        </ol>
        <Text>
            See {/* TODO: proper docs link here */}
            <Link rel="noopener noreferrer" target="_blank" to="">
                the docs for more options
            </Link>
            , or try one of the buttons below.
        </Text>
    </div>
)

const GitLabInstructions: React.FunctionComponent<{ isSelfManaged: boolean }> = ({ isSelfManaged }) => (
    <div>
        <ol>
            {isSelfManaged && (
                <li>
                    Set <Field>url</Field> to the URL of GitLab.
                </li>
            )}
            <li>
                Create a GitLab access token (
                <Link
                    to="https://docs.gitlab.com/ee/user/profile/personal_access_tokens.html#creating-a-personal-access-token"
                    target="_blank"
                    rel="noopener noreferrer"
                >
                    instructions
                </Link>
                ) with{' '}
                <b>
                    <Field>repo</Field>
                </b>{' '}
                scope, and set it to be the value of the <Field>token</Field> field in the configuration below.
            </li>
            <li>
                Use the following fields to select projects:
                <ul>
                    <li>
                        <Field>projectQuery</Field> is a list of calls to{' '}
                        <Link
                            target="_blank"
                            rel="noopener noreferrer"
                            to="https://docs.gitlab.com/ee/api/projects.html"
                        >
                            GitLab's REST API
                        </Link>{' '}
                        that return a list of projects.
                        <br />
                        <Value>"groups/&lt;mygroup&gt;/projects"</Value> selects all projects in a group.
                        <br />
                        <Value>"projects?membership=true&archived=no"</Value> selects all unarchived projects of which
                        the token's user is a member.
                        <br />
                        <Value>"search?scope=projects&search=my_search_query"</Value> selects all projects matching a
                        search query.
                    </li>
                    <li>
                        <Field>projects</Field> is a list of individual projects.
                    </li>
                    <li>
                        <Field>exclude</Field> excludes individual projects.
                    </li>
                </ul>
            </li>
        </ol>
        <Text>
            See{' '}
            <Link rel="noopener noreferrer" target="_blank" to="/help/admin/code_hosts/gitlab#configuration">
                the docs for more options
            </Link>
            , or try one of the buttons below.
        </Text>
    </div>
)

const githubEditorActions = (isEnterprise: boolean): EditorAction[] => [
    ...(isEnterprise
        ? [
              {
                  id: 'setURL',
                  label: 'Set GitHub URL',
                  run: (config: string) => {
                      const value = 'https://github.example.com'
                      const edits = modify(config, ['url'], value, defaultModificationOptions)
                      return { edits, selectText: value }
                  },
              },
          ]
        : []),
    {
        id: 'setAccessToken',
        label: 'Set access token',
        run: (config: string) => {
            const value = '<access token>'
            const edits = modify(config, ['token'], value, defaultModificationOptions)
            return { edits, selectText: '<access token>' }
        },
    },
    {
        id: 'addOrgRepo',
        label: 'Add repositories in an organization',
        run: (config: string) => {
            const value = '<organization name>'
            const edits = modify(config, ['orgs', -1], value, defaultModificationOptions)
            return { edits, selectText: '<organization name>' }
        },
    },
    {
        id: 'addSearchQueryRepos',
        label: 'Add repositories matching a search query',
        run: (config: string) => {
            const value = '<search query>'
            const edits = modify(config, ['repositoryQuery', -1], value, defaultModificationOptions)
            return { edits, selectText: '<search query>' }
        },
    },
    {
        id: 'addAffiliatedRepos',
        label: 'Add repositories affiliated with token',
        run: (config: string) => {
            const value = 'affiliated'
            const edits = modify(config, ['repositoryQuery', -1], value, defaultModificationOptions)
            return { edits, selectText: 'affiliated' }
        },
    },
    {
        id: 'addRepo',
        label: 'Add a single repository',
        run: (config: string) => {
            const value = '<owner>/<repository>'
            const edits = modify(config, ['repos', -1], value, defaultModificationOptions)
            return { edits, selectText: '<owner>/<repository>' }
        },
    },
    {
        id: 'excludeRepo',
        label: 'Exclude a repository',
        run: (config: string) => {
            const value = { name: '<owner>/<repository>' }
            const edits = modify(config, ['exclude', -1], value, defaultModificationOptions)
            return { edits, selectText: '<owner>/<repository>' }
        },
    },
    {
        id: 'enablePermissions',
        label: 'Enforce permissions',
        run: (config: string) => {
            const value = {
                COMMENT_SENTINEL: true,
            }
            const comment = editorActionComments.enablePermissions
            const edit = editWithComment(config, ['authorization'], value, comment)
            return { edits: [edit], selectText: comment }
        },
    },
    {
        id: 'addWebhooks',
        label: 'Add webhook',
        run: (config: string) => {
            const value = { org: '<your_org_on_GitHub>', secret: '<any_secret_string>' }
            const edits = modify(config, ['webhooks', -1], value, defaultModificationOptions)
            return { edits, selectText: '<your_org_on_GitHub>' }
        },
    },
]

const gitHubAppEditorActions = (): EditorAction[] => {
    const actions = githubEditorActions(true).filter(
        item =>
            !['setURL', 'setAccessToken', 'addAffiliatedRepos', 'enablePermissions', 'addWebhooks'].includes(item.id)
    )
    return [
        {
            id: 'addAffiliatedRepos',
            label: 'Add repositories affiliated with app installation',
            run: (config: string) => {
                const value = 'affiliated'
                const edits = modify(config, ['repositoryQuery', -1], value, defaultModificationOptions)
                return { edits, selectText: 'affiliated' }
            },
        },
        ...actions,
    ]
}

const gitlabEditorActions = (isSelfManaged: boolean): EditorAction[] => [
    ...(isSelfManaged
        ? [
              {
                  id: 'setURL',
                  label: 'Set GitLab URL',
                  run: (config: string) => {
                      const value = 'https://gitlab.example.com'
                      const edits = modify(config, ['url'], value, defaultModificationOptions)
                      return { edits, selectText: value }
                  },
              },
          ]
        : []),
    {
        id: 'setAccessToken',
        label: 'Set access token',
        run: (config: string) => {
            const value = '<access token>'
            const edits = modify(config, ['token'], value, defaultModificationOptions)
            return { edits, selectText: value }
        },
    },
    {
        id: 'addGroupProjects',
        label: 'Add projects in a group',
        run: (config: string) => {
            const value = 'groups/<my group>/projects'
            const edits = modify(config, ['projectQuery', -1], value, defaultModificationOptions)
            return { edits, selectText: '<my group>' }
        },
    },
    {
        id: 'addMemberProjects',
        label: "Add projects that have the token's user as member",
        run: (config: string) => {
            const value = 'projects?membership=true&archived=no'
            const edits = modify(config, ['projectQuery', -1], value, defaultModificationOptions)
            return { edits, selectText: value }
        },
    },
    {
        id: 'addProjectsMatchingSearch',
        label: 'Add projects matching search',
        run: (config: string) => ({
            edits: modify(config, ['projectQuery', -1], '?search=<search query>', defaultModificationOptions),
            selectText: '<search query>',
        }),
    },
    {
        id: 'addIndividualProjectByName',
        label: 'Add single project by name',
        run: (config: string) => {
            const value = { name: '<group>/<name>' }
            const edits = modify(config, ['projects', -1], value, defaultModificationOptions)
            return { edits, selectText: '<group>/<name>' }
        },
    },
    {
        id: 'addIndividualProjectByID',
        label: 'Add single project by ID',
        run: (config: string) => {
            const value = { id: 123 }
            const edits = modify(config, ['projects', -1], value, defaultModificationOptions)
            return { edits, selectText: '123' }
        },
    },
    ...(isSelfManaged
        ? [
              {
                  id: 'addInternalProjects',
                  label: 'Add internal projects',
                  run: (config: string) => {
                      const value = 'projects?visibility=internal'
                      const edits = modify(config, ['projectQuery', -1], value, defaultModificationOptions)
                      return { edits, selectText: value }
                  },
              },
              {
                  id: 'addPrivateProjects',
                  label: 'Add private projects',
                  run: (config: string) => {
                      const value = 'projects?visibility=private'
                      const edits = modify(config, ['projectQuery', -1], value, defaultModificationOptions)
                      return { edits, selectText: value }
                  },
              },
          ]
        : []),
    {
        id: 'excludeProject',
        label: 'Exclude a project',
        run: (config: string) => {
            const value = { name: '<group>/<project>' }
            const edits = modify(config, ['exclude', -1], value, defaultModificationOptions)
            return { edits, selectText: '"<group>/<project>"' }
        },
    },
    ...(isSelfManaged
        ? [
              {
                  id: 'enforcePermissionsOAuth',
                  label: 'Enforce permissions (OAuth)',
                  run: (config: string) => {
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
                  id: 'enforcePermissionsSudo',
                  label: 'Enforce permissions (sudo)',
                  run: (config: string) => {
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
                  id: 'setSelfSignedCert',
                  label: 'Set internal or self-signed certificate',
                  run: (config: string) => {
                      const value = '<certificate>'
                      const edits = modify(config, ['certificate'], value, defaultModificationOptions)
                      return { edits, selectText: value }
                  },
              },
          ]
        : [
              {
                  id: 'enforcePermissionsOAuth',
                  label: 'Enforce permissions',
                  run: (config: string) => {
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
          ]),
    {
        id: 'addWebhooks',
        label: 'Add webhook',
        run: (config: string) => {
            const value = { secret: '<any_secret_string>' }
            const edits = modify(config, ['webhooks', -1], value, defaultModificationOptions)
            return { edits, selectText: '<any_secret_string>' }
        },
    },
]

const azureDevOpsEditorActions = (): EditorAction[] => [
    {
        id: 'excludeRepo',
        label: 'Exclude a repository',
        run: (config: string) => {
            const value = { name: '<project>/<repository>' }
            const edits = modify(config, ['exclude', -1], value, defaultModificationOptions)
            return { edits, selectText: '<project>/<repository>' }
        },
    },
]

export const GITHUB: AddExternalServiceOptions = {
    kind: ExternalServiceKind.GITHUB,
    shortDescription: 'Cloud-hosted or Self-hosted using an access token',
    title: 'GitHub.com or Enterprise',
    icon: GithubIcon,
    jsonSchema: githubSchemaJSON,
    editorActions: githubEditorActions(false),
    Instructions: () => <GitHubInstructions isEnterprise={false} />,
    defaultDisplayName: 'GitHub',
    defaultConfig: `{
  "url": "https://github.example.com",
  "token": "<access token>",
  "orgs": []
}`,
}
const GITHUB_APP: AddExternalServiceOptions = {
    ...GITHUB,
    title: 'GitHub App',
    shortDescription:
        'Connect repositories on GitHub.com or self-hosted GitHub Enterprise installations using a GitHub App installation',
    editorActions: gitHubAppEditorActions(),
    Instructions: () => <GitHubAppInstructions />,
    additionalFormComponent: <GitHubAppSelector />,
}
export const gitHubAppConfig = (
    baseURL: string | null,
    appID: string | null,
    installationID: string | null
): AddExternalServiceOptions => ({
    ...GITHUB_APP,
    defaultConfig: `{
  "url": "${baseURL ? decodeURI(baseURL) : ''}",
  "gitHubAppDetails": {
    "installationID": ${installationID ?? -1},
    "appID": ${appID ?? -1},
    "baseURL": "${baseURL ?? ''}",
    "cloneAllRepositories": true
  },
  "authorization": {}
}`,
})

const AWS_CODE_COMMIT: AddExternalServiceOptions = {
    kind: ExternalServiceKind.AWSCODECOMMIT,
    title: 'AWS CodeCommit repositories',
    icon: AwsIcon,
    jsonSchema: awsCodeCommitSchemaJSON,
    defaultDisplayName: 'AWS CodeCommit',
    defaultConfig: `{
  "accessKeyID": "<access key id>",
  "secretAccessKey": "<secret access key>",
  "region": "<region>",
  "gitCredentials": {
    "username": "<username>",
    "password": "<password>"
  }
}`,
    Instructions: () => (
        <div>
            <ol>
                <li>
                    Obtain your AWS Secret Access Key and key ID:
                    <ul>
                        <li>Sign in to your AWS Management Console.</li>
                        <li>Click your username in the upper right corner of the page.</li>
                        <li>Click on "My Security Credentials".</li>
                        <li>Scroll to the section, "Access keys for CLI, SDK, & API access".</li>
                        <li>
                            Use an existing key ID if you still have access to its Secret Access Key. Otherwise, click
                            "Create access key" to create a new key. Record the Secret Access Key and key ID in a safe
                            place.
                        </li>
                        <li>
                            Set <Field>accessKeyID</Field> and <Field>secretAccessKey</Field> in the configuration below
                            to the access key ID and Secret Access Key.
                        </li>
                    </ul>
                </li>
                <li>
                    Set the region to your AWS region. The region (e.g., <Value>us-west-2</Value>) should be visible in
                    the URL when you visit AWS CodeCommit. You can visit AWS CodeCommit by logging into AWS, clicking on
                    "Services" in the top navbar, and clicking on "CodeCommit".
                </li>
                <li>
                    Create Git credentials for AWS CodeCommit (
                    <Link
                        to="https://docs.aws.amazon.com/codecommit/latest/userguide/setting-up-gc.html#setting-up-gc-iam"
                        target="_blank"
                        rel="noopener noreferrer"
                    >
                        instructions
                    </Link>
                    ) and set these in the <Field>gitCredentials</Field> field.
                </li>
                <li>
                    You can optionally exclude repositories using the <Field>exclude</Field> field.
                </li>
            </ol>
            <Text>
                See{' '}
                <Link
                    rel="noopener noreferrer"
                    target="_blank"
                    to="/help/admin/code_hosts/aws_codecommit#configuration"
                >
                    the docs for more options
                </Link>
                , or try one of the buttons below.
            </Text>
        </div>
    ),
    editorActions: [
        {
            id: 'setAccessKeyID',
            label: 'Set access key ID',
            run: (config: string) => {
                const value = '<access key id>'
                const edits = modify(config, ['accessKeyID'], value, defaultModificationOptions)
                return { edits, selectText: value }
            },
        },
        {
            id: 'setSecretAccessKey',
            label: 'Set secret access key',
            run: (config: string) => {
                const value = '<secret access key>'
                const edits = modify(config, ['secretAccessKey'], value, defaultModificationOptions)
                return { edits, selectText: value }
            },
        },
        {
            id: 'setRegion',
            label: 'Set region',
            run: (config: string) => {
                const value = '<region>'
                const edits = modify(config, ['region'], value, defaultModificationOptions)
                return { edits, selectText: value }
            },
        },
        {
            id: 'setGitCredentials',
            label: 'Set Git credentials',
            run: (config: string) => {
                const value = {
                    username: '<username>',
                    password: '<password>',
                }
                const edits = modify(config, ['gitCredentials'], value, defaultModificationOptions)
                return { edits, selectText: '<username>' }
            },
        },
        {
            id: 'excludeRepo',
            label: 'Exclude a repository',
            run: (config: string) => {
                const value = { name: '<owner>/<repository>' }
                const edits = modify(config, ['exclude', -1], value, defaultModificationOptions)
                return { edits, selectText: '<owner>/<repository>' }
            },
        },
    ],
}
const BITBUCKET_CLOUD: AddExternalServiceOptions = {
    kind: ExternalServiceKind.BITBUCKETCLOUD,
    title: 'Bitbucket.org',
    icon: BitbucketIcon,
    jsonSchema: bitbucketCloudSchemaJSON,
    defaultDisplayName: 'Bitbucket Cloud',
    defaultConfig: `{
  "url": "https://bitbucket.org",
  "appPassword": "<app password>",
  "username": "<username to which the app password belongs>",
  "teams": []
}`,
    editorActions: [
        {
            id: 'setAppPassword',
            label: 'Set app password',
            run: (config: string) => {
                const value = '<app password>'
                const edits = modify(config, ['appPassword'], value, defaultModificationOptions)
                return { edits, selectText: value }
            },
        },
        {
            id: 'setUsername',
            label: 'Set username',
            run: (config: string) => {
                const value = '<username to which the app password belongs>'
                const edits = modify(config, ['username'], value, defaultModificationOptions)
                return { edits, selectText: value }
            },
        },
        {
            id: 'addTeamRepositories',
            label: 'Add repositories belonging to team',
            run: (config: string) => {
                const value = '<team>'
                const edits = modify(config, ['teams', -1], value, defaultModificationOptions)
                return { edits, selectText: value }
            },
        },
        {
            id: 'enableWebhooks',
            label: 'Enable webhooks',
            run: (config: string) => {
                const value = '<any_secret_string>'
                const edits = modify(config, ['webhookSecret'], value, defaultModificationOptions)
                return { edits, selectText: value }
            },
        },
    ],
    Instructions: () => (
        <div>
            <ol>
                <li>
                    Create a Bitbucket app password (
                    <Link
                        to="https://confluence.atlassian.com/bitbucket/app-passwords-828781300.html"
                        target="_blank"
                        rel="noopener noreferrer"
                    >
                        instructions
                    </Link>
                    ) with the following scopes:
                    <ul>
                        <li>
                            <strong>Workspace membership</strong>: Read
                        </li>
                        <li>
                            <strong>Repositories</strong>: Read
                            <ul>
                                <li>
                                    To enable repository permissions syncing, <strong>Repositories</strong>: Admin is
                                    required.
                                </li>
                            </ul>
                        </li>
                    </ul>
                </li>
                <li>
                    Paste your app password in the <Field>appPassword</Field> field in the configuration below.
                </li>
                <li>
                    Set the <Field>username</Field> field to be the username corresponding to <Field>appPassword</Field>
                    .
                </li>
                <li>
                    Set the <Field>teams</Field> field to be the list of teams whose repositories Sourcegraph should
                    index.
                </li>
            </ol>
            <Text>
                See{' '}
                <Link
                    rel="noopener noreferrer"
                    target="_blank"
                    to="/help/admin/code_hosts/bitbucket_cloud#configuration"
                >
                    the docs for more options
                </Link>
                , or try one of the buttons below.
            </Text>
        </div>
    ),
    shortDescription: 'Cloud-hosted using credentials',
}
const BITBUCKET_SERVER: AddExternalServiceOptions = {
    kind: ExternalServiceKind.BITBUCKETSERVER,
    title: 'Bitbucket Server',
    icon: BitbucketIcon,
    jsonSchema: bitbucketServerSchemaJSON,
    defaultDisplayName: 'Bitbucket Server',
    defaultConfig: `{
  "url": "https://bitbucket.example.com",
  "token": "<access token>",
  "username": "<username that created access token>",
  "repositoryQuery": [
    "all"
  ]
}`,
    Instructions: () => (
        <div>
            <ol>
                <li>
                    In the configuration below, set <Field>url</Field> to the URL of Bitbucket Server.
                </li>
                <li>
                    Create a personal access token (
                    <Link
                        to="https://confluence.atlassian.com/bitbucketserver/personal-access-tokens-939515499.html"
                        target="_blank"
                        rel="noopener noreferrer"
                    >
                        instructions
                    </Link>
                    ) with <Field>read</Field> scope.
                </li>
                <li>
                    Set <Field>username</Field> to the user that created the personal access token.
                </li>
                <li>
                    Specify which repositories Sourcegraph should clone using the following fields.
                    <ul>
                        <li>
                            <Field>repositoryQuery</Field>: a list of strings that are one of the following:
                            <ul>
                                <li>
                                    <Value>"all"</Value> selects all repositories visible to the token
                                </li>
                                <li>
                                    A query string like{' '}
                                    <Value>"{'?name=<repo name>&projectname=<project>&visibility=private'}"</Value> that
                                    specifies search query parameters. See{' '}
                                    <Link
                                        to="https://docs.atlassian.com/bitbucket-server/rest/6.1.2/bitbucket-rest.html#idp355"
                                        target="_blank"
                                        rel="noopener noreferrer"
                                    >
                                        the full list of parameters
                                    </Link>
                                    .
                                </li>
                                <li>
                                    <Value>"none"</Value> selects no repositories (should only be used if you are
                                    listing repositories one-by-one)
                                </li>
                            </ul>
                        </li>
                        <li>
                            <Field>repos</Field>: a list of single repositories
                        </li>
                        <li>
                            <Field>exclude</Field>: a list of repositories or repository name patterns to exclude
                        </li>
                        <li>
                            <Field>excludePersonalRepositories</Field>: if true, excludes personal repositories from
                            being indexed
                        </li>
                    </ul>
                </li>
            </ol>
            <Text>
                See{' '}
                <Link
                    rel="noopener noreferrer"
                    target="_blank"
                    to="/help/admin/code_hosts/bitbucket_server#configuration"
                >
                    the docs for more options
                </Link>
                , or try one of the buttons below.
            </Text>
        </div>
    ),
    editorActions: [
        {
            id: 'setURL',
            label: 'Set URL',
            run: (config: string) => {
                const value = 'https://bitbucket.example.com'
                const edits = modify(config, ['url'], value, defaultModificationOptions)
                return { edits, selectText: value }
            },
        },
        {
            id: 'setPersonalAccessToken',
            label: 'Set access token',
            run: (config: string) => {
                const value = '<access token>'
                const edits = modify(config, ['token'], value, defaultModificationOptions)
                return { edits, selectText: value }
            },
        },
        {
            id: 'setUsername',
            label: 'Set username',
            run: (config: string) => {
                const value = '<username that created access token>'
                const edits = modify(config, ['username'], value, defaultModificationOptions)
                return { edits, selectText: value }
            },
        },
        {
            id: 'addProjectRepos',
            label: 'Add repositories in a project',
            run: (config: string) => {
                const value = '?projectname=<project>'
                const edits = modify(config, ['repositoryQuery', -1], value, defaultModificationOptions)
                return { edits, selectText: '<project>' }
            },
        },
        {
            id: 'addRepo',
            label: 'Add individual repository',
            run: (config: string) => {
                const value = '<project/<repository>'
                const edits = modify(config, ['repos', -1], value, defaultModificationOptions)
                return { edits, selectText: '<project/<repository>' }
            },
        },
        {
            id: 'excludeRepo',
            label: 'Exclude a repository',
            run: (config: string) => {
                const value = { name: '<project/<repository>' }
                const edits = modify(config, ['exclude', -1], value, defaultModificationOptions)
                return { edits, selectText: '{"name": "<project/<repository>"}' }
            },
        },
        {
            id: 'setSelfSignedCert',
            label: 'Set internal or self-signed certificate',
            run: (config: string) => {
                const value = '<certificate>'
                const edits = modify(config, ['certificate'], value, defaultModificationOptions)
                return { edits, selectText: value }
            },
        },
        {
            id: 'enableWebhooks',
            label: 'Enable webhooks',
            run: (config: string) => {
                const value = { webhooks: { secret: '<any_secret_string>' } }
                const edits = modify(config, ['plugin'], value, defaultModificationOptions)
                return { edits, selectText: '<any_secret_string>' }
            },
        },
    ],
    shortDescription: 'Self-hosted using an access token',
}
const GITLAB: AddExternalServiceOptions = {
    kind: ExternalServiceKind.GITLAB,
    title: 'GitLab',
    icon: GitLabIcon,
    shortDescription: 'Connect cloud-hosted or self-hosted repositories on GitLab',
    jsonSchema: gitlabSchemaJSON,
    defaultDisplayName: 'GitLab',
    defaultConfig: `{
  "url": "https://gitlab.example.com",
  "token": "<access token>",
  "projectQuery": [
    "projects?membership=true&archived=no"
  ]
}`,
    editorActions: gitlabEditorActions(false),
    Instructions: () => <GitLabInstructions isSelfManaged={false} />,
}
const SRC_SERVE_GIT: AddExternalServiceOptions = {
    kind: ExternalServiceKind.OTHER,
    title: 'Sourcegraph CLI Serve-Git',
    icon: GitIcon,
    jsonSchema: otherExternalServiceSchemaJSON,
    defaultDisplayName: 'src serve-git',
    defaultConfig: `{
  // url is the http url to 'src serve-git'.
  // url should be reachable by Sourcegraph.
  "url": "http://addr.for.src.serve:3434",

  // Do not change this. Sourcegraph uses this as a signal that url is 'src serve'.
  "repos": ["src-serve"]
}`,
    Instructions: () => (
        <div>
            <Text>
                In the configuration below, set <Field>url</Field> to be the URL of src serve-git.
            </Text>
            <Text>
                Install the{' '}
                <Link rel="noopener noreferrer" target="_blank" to="https://github.com/sourcegraph/src-cli">
                    Sourcegraph CLI (src)
                </Link>
                . src serve-git allows you to serve any git repositories that you have on disk.
            </Text>
        </div>
    ),
    editorActions: [
        {
            id: 'setURL',
            label: 'Sourcegraph in Docker and src serve-git running on host',
            run: (config: string) => {
                const value = 'http://host.docker.internal:3434'
                const edits = modify(config, ['url'], value, defaultModificationOptions)
                return { edits, selectText: value }
            },
        },
    ],
}
const GITOLITE: AddExternalServiceOptions = {
    kind: ExternalServiceKind.GITOLITE,
    title: 'Gitolite',
    icon: GitIcon,
    jsonSchema: gitoliteSchemaJSON,
    defaultDisplayName: 'Gitolite',
    defaultConfig: `{
  "host": "git@gitolite.example.com",
  "prefix": "gitolite.example.com/"
}`,
    Instructions: () => (
        <div>
            <ol>
                <li>
                    In the configuration below, set <Field>host</Field> to be the username and host of the Gitolite
                    server.
                </li>
                <li>
                    Set the <Field>prefix</Field> field to the prefix you desire for the repository names on
                    Sourcegraph. This is typically the hostname of the Gitolite server.
                </li>
            </ol>
            <Text>
                See{' '}
                <Link rel="noopener noreferrer" target="_blank" to="/help/admin/code_hosts/gitolite#configuration">
                    the docs for more advanced options
                </Link>
                , or try one of the buttons below.
            </Text>
        </div>
    ),
    editorActions: [
        {
            id: 'setHost',
            label: 'Set host',
            run: (config: string) => {
                const value = 'git@gitolite.example.com'
                const edits = modify(config, ['host'], value, defaultModificationOptions)
                return { edits, selectText: value }
            },
        },
        {
            id: 'setPrefix',
            label: 'Set prefix',
            run: (config: string) => {
                const value = 'gitolite.example.com/'
                const edits = modify(config, ['prefix'], value, defaultModificationOptions)
                return { edits, selectText: value }
            },
        },
        {
            id: 'excludeRepo',
            label: 'Exclude a repository',
            run: (config: string) => {
                const value = { name: '<name>' }
                const edits = modify(config, ['exclude', -1], value, defaultModificationOptions)
                return { edits, selectText: '<name>' }
            },
        },
    ],
}
const PHABRICATOR_SERVICE: AddExternalServiceOptions = {
    kind: ExternalServiceKind.PHABRICATOR,
    title: 'Phabricator connection',
    icon: PhabricatorIcon,
    shortDescription:
        'Associate Phabricator repositories with existing repositories on Sourcegraph. Mirroring is not supported.',
    jsonSchema: phabricatorSchemaJSON,
    defaultDisplayName: 'Phabricator',
    defaultConfig: `{
  // Use Ctrl+Space for completion, and hover over JSON properties for documentation.
  // Configuration options are documented here:
  // https://sourcegraph.com/docs/admin/code_hosts/phabricator#configuration

  "url": "https://phabricator.example.com",
  "token": "",
  "repos": []
}`,
    editorActions: [
        {
            id: 'setPhabricatorURL',
            label: 'Set Phabricator URL',
            run: (config: string) => {
                const value = 'https://phabricator.example.com'
                const edits = modify(config, ['url'], value, defaultModificationOptions)
                return { edits, selectText: value }
            },
        },
        {
            id: 'setAccessToken',
            label: 'Set Phabricator access token',
            run: (config: string) => {
                const value = '<Phabricator access token>'
                const edits = modify(config, ['token'], value, defaultModificationOptions)
                return { edits, selectText: value }
            },
        },
        {
            id: 'addRepository',
            label: 'Add a repository',
            run: (config: string) => {
                const value = {
                    callsign: '<Phabricator repository callsign>',
                    path: '<Sourcegraph repository full name>',
                }
                const edits = modify(config, ['repos', -1], value, defaultModificationOptions)
                return { edits, selectText: '<Phabricator repository callsign>' }
            },
        },
    ],
}
const GENERIC_GIT: AddExternalServiceOptions = {
    kind: ExternalServiceKind.OTHER,
    title: 'Generic Git host',
    icon: GitIcon,
    jsonSchema: otherExternalServiceSchemaJSON,
    defaultDisplayName: 'Git repositories',
    defaultConfig: `{
  "url": "https://git.example.com",
  "repos": []
}`,
    Instructions: () => (
        <div>
            <ol>
                <li>
                    In the configuration below, set <Field>url</Field> to be the URL of your Git host.
                </li>
                <li>
                    Add the paths of the repositories you wish to index to the <Field>repos</Field> field. These will be
                    appended to the host URL to obtain the repository clone URLs.
                </li>
            </ol>
            <Text>
                See{' '}
                <Link rel="noopener noreferrer" target="_blank" to="/help/admin/code_hosts/other#configuration">
                    the docs for more options
                </Link>
                , or try one of the buttons below.
            </Text>
        </div>
    ),
    editorActions: [
        {
            id: 'setURL',
            label: 'Set Git host URL',
            run: (config: string) => {
                const value = 'https://git.example.com'
                const edits = modify(config, ['url'], value, defaultModificationOptions)
                return { edits, selectText: value }
            },
        },
        {
            id: 'addRepo',
            label: 'Add a repository',
            run: (config: string) => {
                const value = 'path/to/repository'
                const edits = modify(config, ['repos', -1], value, defaultModificationOptions)
                return { edits, selectText: value }
            },
        },
    ],
}

const PERFORCE: AddExternalServiceOptions = {
    kind: ExternalServiceKind.PERFORCE,
    title: 'Perforce',
    icon: PerforceIcon,
    jsonSchema: perforceSchemaJSON,
    defaultDisplayName: 'Perforce',
    defaultConfig: `{
  "p4.port": "ssl:111.222.333.444:1666",
  "p4.user": "admin",
  "p4.passwd": "<ticket value>",
  "depots": []
}`,
    Instructions: () => (
        <div>
            <ol>
                <li>
                    In the configuration below, set <Field>p4.port</Field> to be the Perforce Server address.
                </li>
                <li>
                    Set the <Field>p4.user</Field> field to be the authenticated user.
                </li>
                <li>
                    Set the <Field>p4.passwd</Field> field to be the ticket value of the authenticated user.
                </li>
            </ol>
            <Text>
                See{' '}
                <Link rel="noopener noreferrer" target="_blank" to="/help/admin/repo/perforce#configuration">
                    the docs for more advanced options
                </Link>
                , or try one of the buttons below.
            </Text>
        </div>
    ),
    editorActions: [
        {
            id: 'setMaxChanges',
            label: 'Set max changes',
            run: (config: string) => {
                const value = 1000
                const edits = modify(config, ['maxChanges'], value, defaultModificationOptions)
                return { edits, selectText: value }
            },
        },
        {
            id: 'enforcePermissions',
            label: 'Enforce permissions',
            run: (config: string) => {
                const value = {}
                const edits = modify(config, ['authorization'], value, defaultModificationOptions)
                return { edits, selectText: '"authorization": {}' }
            },
        },
    ],
}
const JVM_PACKAGES: AddExternalServiceOptions = {
    kind: ExternalServiceKind.JVMPACKAGES,
    title: 'JVM Dependencies',
    icon: LanguageJavaIcon,
    jsonSchema: jvmPackagesSchemaJSON,
    defaultDisplayName: 'JVM Dependencies',
    defaultConfig: `{
  "maven": {
    "repositories": [],
    "dependencies": []
  }
}`,
    Instructions: () => (
        <div>
            <ol>
                <li>
                    In the configuration below, set <Field>maven.repositories</Field> to the list of Maven repositories.
                    For example,
                    <Code>"https://maven.google.com"</Code>.
                </li>
                <li>
                    In the configuration below, set <Field>maven.dependencies</Field> to the list of artifacts that you
                    want to manually add. For example,
                    <Code>"junit:junit:4.13.2"</Code> or
                    <Code>"org.hamcrest:hamcrest-core:1.3:default"</Code>.
                </li>
            </ol>
            <Text>⚠️ JVM dependency repositories are visible by all users of the Sourcegraph instance.</Text>
            <Text>⚠️ It is only possible to register one JVM dependency code host per Sourcegraph instance.</Text>
        </div>
    ),
    editorActions: [],
}

const PAGURE: AddExternalServiceOptions = {
    kind: ExternalServiceKind.PAGURE,
    title: 'Pagure',
    icon: GitIcon,
    jsonSchema: pagureSchemaJSON,
    defaultDisplayName: 'Pagure',
    defaultConfig: `{
  "url": "https://pagure.example.com",
}`,
    Instructions: () => (
        <div>
            <ol>
                <li>
                    In the configuration below, set <Field>url</Field> to the URL of Pagure instance.
                </li>
            </ol>
        </div>
    ),
    editorActions: [],
}

const GERRIT: AddExternalServiceOptions = {
    kind: ExternalServiceKind.GERRIT,
    title: 'Gerrit',
    icon: GerritIcon,
    jsonSchema: gerritSchemaJSON,
    defaultDisplayName: 'Gerrit',
    defaultConfig: `{
  "url": "https://gerrit.example.com",
  "username": "<fill in>",
  "password": "<fill in>"
}`,
    Instructions: () => (
        <div>
            <ol>
                <li>
                    In the configuration below, set <Field>url</Field> to the URL of Gerrit instance.
                </li>
                <li>
                    Provide <Field>username</Field> and <Field>password</Field> of a user account that has access to all
                    the repositories you want to sync.
                </li>
                <li>
                    Optionally, use <Field>projects</Field> to limit syncing from the Gerrit instance to specific
                    projects.
                </li>
            </ol>
        </div>
    ),
    editorActions: [],
    status: 'beta',
}

const AZUREDEVOPS: AddExternalServiceOptions = {
    kind: ExternalServiceKind.AZUREDEVOPS,
    title: 'Azure DevOps',
    icon: AzureDevOpsIcon,
    jsonSchema: azureDevOpsSchemaJSON,
    defaultDisplayName: 'Azure DevOps',
    defaultConfig: `{
  "url": "https://dev.azure.com",
  "username": "<username>",
  "token": "<token>",
  "orgs": [],
  "projects": []
}`,
    Instructions: () => (
        <div>
            <ol>
                <li>
                    In the configuration below, set <Field>url</Field> to the URL of Azure DevOps Services:{' '}
                    <Link to="https://dev.azure.com">https://dev.azure.com</Link>.
                </li>
                <li>
                    In the configuration below, set <Field>username</Field> to the authenticated username for Azure
                    DevOps Services.
                </li>
                <li>
                    In the configuration below, set <Field>token</Field> to the authenticated token for Azure DevOps
                    Services. See the{' '}
                    <Link to="https://learn.microsoft.com/en-us/azure/devops/organizations/accounts/use-personal-access-tokens-to-authenticate?view=azure-devops&tabs=Windows#create-a-pat">
                        Azure DevOps documentation
                    </Link>{' '}
                    for instructions on how to create a Personal Access Token.
                </li>
                <li>
                    In the configuration below, set <Field>orgs</Field> and/or <Field>projects</Field> to the
                    organizations/projects you want to sync repositories from.
                </li>
            </ol>
        </div>
    ),
    editorActions: azureDevOpsEditorActions(),
}

const NPM_PACKAGES: AddExternalServiceOptions = {
    kind: ExternalServiceKind.NPMPACKAGES,
    title: 'npm Dependencies',
    icon: NpmIcon,
    jsonSchema: npmPackagesSchemaJSON,
    defaultDisplayName: 'npm Dependencies',
    defaultConfig: `{
  "registry": "https://registry.npmjs.org",
  "dependencies": []
}`,
    Instructions: () => (
        <div>
            <ol>
                <li>
                    In the configuration below, set <Field>registry</Field> to the applicable npm registry. For example,
                    <Code>"https://registry.npmjs.mycompany.com"</Code> or <Code>"https://registry.npmjs.org"</Code>.
                    Note that this URL may not be the same as where packages can be searched (such as{' '}
                    <Code>https://www.npmjs.org</Code>). If you're unsure about the exact URL URL for a custom registry,
                    check the URLs for packages that have already been resolved, such as those in existing lock files
                    like <Code>yarn.lock</Code>.
                </li>
                <li>
                    In the configuration below, set <Field>dependencies</Field> to the list of packages that you want to
                    manually add. For example,
                    <Code>"react@17.0.2"</Code> or <Code>"@types/lodash@4.14.177"</Code>. Version ranges are not
                    supported.
                </li>
            </ol>
            <Text>⚠️ npm package repositories are visible by all users of the Sourcegraph instance.</Text>
            <Text>⚠️ It is only possible to register one npm package code host per Sourcegraph instance.</Text>
        </div>
    ),
    editorActions: [],
}

const GO_MODULES: AddExternalServiceOptions = {
    kind: ExternalServiceKind.GOMODULES,
    title: 'Go Dependencies',
    icon: LanguageGoIcon,
    jsonSchema: goModulesSchemaJSON,
    defaultDisplayName: 'Go Dependencies',
    defaultConfig: `{
  "urls": ["https://proxy.golang.org"],
  "dependencies": []
}`,
    Instructions: () => (
        <div>
            <ol>
                <li>
                    In the configuration below, set <Field>urls</Field> to the Go module proxies you want to sync
                    dependency repositories from. For example, <Code>"https://user:pass@athens.mycompany.com"</Code> or{' '}
                    <Code>"https://proxy.golang.org"</Code>. A module will be synced from the first proxy that has it,
                    trying the next when it's not found.
                </li>
                <li>
                    In the configuration below, set <Field>dependencies</Field> to the list of packages that you want to
                    manually add. For example, <Code>"cloud.google.com/go/kms@v1.1.0"</Code>.
                </li>
            </ol>
            <Text>⚠️ go module repositories are visible by all users of the Sourcegraph instance.</Text>
            <Text>⚠️ It is only possible to register one go modules code host per Sourcegraph instance.</Text>
        </div>
    ),
    editorActions: [],
}

const PYTHON_PACKAGES: AddExternalServiceOptions = {
    kind: ExternalServiceKind.PYTHONPACKAGES,
    title: 'Python Dependencies',
    icon: LanguagePythonIcon,
    jsonSchema: pythonPackagesJSON,
    defaultDisplayName: 'Python Dependencies',
    defaultConfig: `{
  "urls": ["https://pypi.org/simple"],
  "dependencies": []
}`,
    Instructions: () => (
        <div>
            <ol>
                <li>
                    In the configuration below, set <Field>urls</Field> to the simple repository APIs you want to sync
                    dependency repositories from. For example,{' '}
                    <Code>"https://user:pass@artifactory.mycompany.com/simple"</Code> or{' '}
                    <Code>"https://pypi.org/simple"</Code>. A package will be synced from the first API that has it,
                    trying the next when it's not found.
                </li>
                <li>
                    In the configuration below, set <Field>dependencies</Field> to the list of packages that you want to
                    manually add. For example, <Code>"numpy==1.22.3"</Code>.
                </li>
            </ol>
            <Text>⚠️ Python package repositories are visible by all users of the Sourcegraph instance.</Text>
            <Text>⚠️ It is only possible to register one Python packages code host per Sourcegraph instance.</Text>
        </div>
    ),
    editorActions: [],
}

const RUST_PACKAGES: AddExternalServiceOptions = {
    kind: ExternalServiceKind.RUSTPACKAGES,
    title: 'Rust Dependencies',
    icon: LanguageRustIcon,
    jsonSchema: rustPackagesJSON,
    defaultDisplayName: 'Rust Dependencies',
    defaultConfig: `{
  "dependencies": []
}`,
    Instructions: () => (
        <div>
            <ol>
                <li>
                    In the configuration below, set <Field>dependencies</Field> to the list of packages that you want to
                    manually add. For example, <Code>"tokio@18.0.0"</Code>.
                </li>
            </ol>
            <Text>⚠️ Rust package repositories are visible by all users of the Sourcegraph instance.</Text>
            <Text>⚠️ It is only possible to register one Rust packages code host per Sourcegraph instance.</Text>
        </div>
    ),
    editorActions: [],
}

const RUBY_PACKAGES: AddExternalServiceOptions = {
    kind: ExternalServiceKind.RUBYPACKAGES,
    title: 'Ruby Dependencies',
    icon: LanguageRubyIcon,
    jsonSchema: rubyPackagesSchemaJSON,
    defaultDisplayName: 'Ruby Dependencies',
    defaultConfig: `{
  "repository": "https://rubygems.org/",
  "dependencies": ["shopify_api@12.0.0"]
}`,
    Instructions: () => (
        <div>
            <ol>
                <li>
                    The URL https://rubygems.org/ is used if the field
                    <Code>"repository"</Code> is empty.
                </li>
                <li>
                    Use the syntax <Code>"GEM_NAME@GEM_VERSION"</Code> to list a dependency for the{' '}
                    <Code>"dependencies"</Code> field.
                </li>
                <li>
                    The field <Code>"repository"</Code> is redacted because it can include <Code>admin:password</Code>{' '}
                    credentials.
                </li>
                <li>
                    See the{' '}
                    <Link to="https://www.jfrog.com/confluence/display/JFROG/RubyGems+Repositories">
                        Artifactory documentation for RubyGem repositories
                    </Link>{' '}
                    for details on how to configure an internal Artifactory repository.
                </li>
            </ol>
            <Text>⚠️ Ruby package repositories are visible by all users of the Sourcegraph instance.</Text>
            <Text>⚠️ It is only possible to register one Ruby packages code host per Sourcegraph instance.</Text>
        </div>
    ),
    editorActions: [],
}

export const codeHostExternalServices: Record<string, AddExternalServiceOptions> = {
    github: GITHUB,
    ghapp: GITHUB_APP,
    gitlabcom: GITLAB,
    bitbucket: BITBUCKET_CLOUD,
    bitbucketserver: BITBUCKET_SERVER,
    aws_codecommit: AWS_CODE_COMMIT,
    srcservegit: SRC_SERVE_GIT,
    gitolite: GITOLITE,
    git: GENERIC_GIT,
    gerrit: GERRIT,
    azuredevops: AZUREDEVOPS,
    phabricator: PHABRICATOR_SERVICE,
    ...(window.context?.experimentalFeatures?.perforce !== 'disabled' ? { perforce: PERFORCE } : {}),
    ...(window.context?.experimentalFeatures?.pagure === 'enabled' ? { pagure: PAGURE } : {}),
}

export const nonCodeHostExternalServices: Record<string, AddExternalServiceOptions> = {
    ...(window.context?.experimentalFeatures?.pythonPackages === 'enabled' ? { pythonPackages: PYTHON_PACKAGES } : {}),
    ...(window.context?.experimentalFeatures?.rustPackages === 'enabled' ? { rustPackages: RUST_PACKAGES } : {}),
    ...(window.context?.experimentalFeatures?.rubyPackages === 'enabled' ? { rubyPackages: RUBY_PACKAGES } : {}),
    ...(window.context?.experimentalFeatures?.goPackages === 'enabled' ? { goModules: GO_MODULES } : {}),
    ...(window.context?.experimentalFeatures?.jvmPackages === 'enabled' ? { jvmPackages: JVM_PACKAGES } : {}),
    ...(window.context?.experimentalFeatures?.npmPackages === 'enabled' ? { npmPackages: NPM_PACKAGES } : {}),
}

export const allExternalServices = {
    ...codeHostExternalServices,
    ...nonCodeHostExternalServices,
}

export const defaultExternalServices: Record<ExternalServiceKind, AddExternalServiceOptions> = {
    [ExternalServiceKind.GITHUB]: GITHUB,
    [ExternalServiceKind.AZUREDEVOPS]: AZUREDEVOPS,
    [ExternalServiceKind.BITBUCKETCLOUD]: BITBUCKET_CLOUD,
    [ExternalServiceKind.BITBUCKETSERVER]: BITBUCKET_SERVER,
    [ExternalServiceKind.GITLAB]: GITLAB,
    [ExternalServiceKind.GITOLITE]: GITOLITE,
    [ExternalServiceKind.PHABRICATOR]: PHABRICATOR_SERVICE,
    [ExternalServiceKind.OTHER]: GENERIC_GIT,
    [ExternalServiceKind.AWSCODECOMMIT]: AWS_CODE_COMMIT,
    [ExternalServiceKind.PERFORCE]: PERFORCE,
    [ExternalServiceKind.GERRIT]: GERRIT,
    [ExternalServiceKind.PAGURE]: PAGURE,
    [ExternalServiceKind.GOMODULES]: GO_MODULES,
    [ExternalServiceKind.JVMPACKAGES]: JVM_PACKAGES,
    [ExternalServiceKind.NPMPACKAGES]: NPM_PACKAGES,
    [ExternalServiceKind.PYTHONPACKAGES]: PYTHON_PACKAGES,
    [ExternalServiceKind.RUSTPACKAGES]: RUST_PACKAGES,
    [ExternalServiceKind.RUBYPACKAGES]: RUBY_PACKAGES,
}

export const externalRepoIcon = (
    externalRepo: Pick<ExternalRepositoryFields, 'serviceType'>
): React.ComponentType<{ className?: string }> | undefined => {
    const externalServiceKind = externalRepo.serviceType.toUpperCase() as ExternalServiceKind
    return defaultExternalServices[externalServiceKind]?.icon ?? undefined
}

export const EXTERNAL_SERVICE_SYNC_RUNNING_STATUSES = new Set<ExternalServiceSyncJobState>([
    ExternalServiceSyncJobState.QUEUED,
    ExternalServiceSyncJobState.PROCESSING,
    ExternalServiceSyncJobState.CANCELING,
])

export const resolveExternalServiceCategory = (
    externalService: ExternalServiceFieldsWithConfig,
    gitHubApp?: GitHubAppByAppIDResult['gitHubAppByAppID']
): AddExternalServiceOptions => {
    let externalServiceCategory = defaultExternalServices[externalService.kind]
    if (externalService && [ExternalServiceKind.GITHUB, ExternalServiceKind.GITLAB].includes(externalService.kind)) {
        let { parsedConfig } = externalService
        if (!parsedConfig) {
            parsedConfig = parseJSONC(externalService.config)
        }
        if (!parsedConfig) {
            return externalServiceCategory
        }
        // if config contains gitHubAppDetails, we should use GitHub App instead
        if (parsedConfig.gitHubAppDetails && gitHubApp) {
            externalServiceCategory = { ...codeHostExternalServices.ghapp }
            externalServiceCategory.additionalFormComponent = (
                <GitHubAppSelector
                    disabled={true}
                    gitHubApp={{
                        ...parsedConfig.gitHubAppDetails,
                        ...gitHubApp,
                    }}
                />
            )
        }
    }
    return externalServiceCategory
}

export function isValidURL(url: string): boolean {
    try {
        new URL(url)
        return true
    } catch {
        return false
    }
}
