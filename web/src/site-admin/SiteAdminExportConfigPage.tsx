import { RouteComponentProps } from 'react-router'
import { fetchAllConfigAndSettings } from './backend'
import * as GQL from '../../../shared/src/graphql/schema'
import React, { useMemo } from 'react'
import { DynamicallyImportedMonacoSettingsEditor } from '../settings/DynamicallyImportedMonacoSettingsEditor'
import awsCodeCommitJSON from '../../../schema/aws_codecommit.schema.json'
import bitbucketCloudSchemaJSON from '../../../schema/bitbucket_cloud.schema.json'
import bitbucketServerSchemaJSON from '../../../schema/bitbucket_server.schema.json'
import githubSchemaJSON from '../../../schema/github.schema.json'
import gitlabSchemaJSON from '../../../schema/gitlab.schema.json'
import gitoliteSchemaJSON from '../../../schema/gitolite.schema.json'
import otherExternalServiceSchemaJSON from '../../../schema/other_external_service.schema.json'
import phabricatorSchemaJSON from '../../../schema/phabricator.schema.json'
import settingsSchemaJSON from '../../../schema/settings.schema.json'
import siteSchemaJSON from '../../../schema/site.schema.json'
import { PageTitle } from '../components/PageTitle'
import { ExternalServiceKind } from '../../../shared/src/graphql/schema'
import { useObservable } from '../util/useObservable'
import { mapValues, values } from 'lodash'
import { Link } from '../../../shared/src/components/Link'
import { withAuthenticatedUser } from '../auth/withAuthenticatedUser'

/**
 * Minimal shape of a JSON Schema. These values are treated as opaque, so more specific types are
 * not needed.
 */
interface JSONSchema {
    $id: string
    definitions?: Record<string, { type: string }>
}

const externalServices: Record<ExternalServiceKind, JSONSchema> = {
    AWSCODECOMMIT: awsCodeCommitJSON,
    BITBUCKETCLOUD: bitbucketCloudSchemaJSON,
    BITBUCKETSERVER: bitbucketServerSchemaJSON,
    GITHUB: githubSchemaJSON,
    GITLAB: gitlabSchemaJSON,
    GITOLITE: gitoliteSchemaJSON,
    OTHER: otherExternalServiceSchemaJSON,
    PHABRICATOR: phabricatorSchemaJSON,
}

const allConfigSchema = {
    $id: 'all.schema.json#',
    allowComments: true,
    additionalProperties: false,
    properties: {
        site: siteSchemaJSON,
        externalServices: {
            type: 'object',
            properties: mapValues(externalServices, schema => ({ type: 'array', items: schema })),
        },
        settings: {
            type: 'object',
            properties: {
                subjects: {
                    type: 'array',
                    items: {
                        type: 'object',
                        properties: {
                            __typename: {
                                type: 'string',
                            },
                            settingsURL: {
                                type: ['string', 'null'],
                            },
                            contents: {
                                ...settingsSchemaJSON,
                                type: ['object', 'null'],
                            },
                        },
                    },
                },
                final: settingsSchemaJSON,
            },
        },
    },
    definitions: values(externalServices)
        .map(schema => schema.definitions)
        .concat([siteSchemaJSON.definitions, settingsSchemaJSON.definitions])
        .reduce((allDefinitions, definitions) => ({ ...allDefinitions, ...definitions }), {}),
}

interface Props extends RouteComponentProps {
    isLightTheme: boolean
    authenticatedUser: GQL.IUser
}

const AuthenticatedSiteAdminExportConfigPage: React.FunctionComponent<Props> = ({
    isLightTheme,
    history,
    authenticatedUser,
}) => {
    const allConfig = useObservable(useMemo(fetchAllConfigAndSettings, []))
    return (
        <div className="site-admin-export-config-page">
            <PageTitle title="Export configuration - Admin" />
            <div className="d-flex justify-content-between align-items-center mt-3 mb-1">
                <h2 className="mb-0">Export configuration and settings</h2>
            </div>
            <p>
                All configuration and settings, except critical site configuration (which must be accessed through the{' '}
                <a href="/help/admin/management_console ">management console</a>).
            </p>
            <div className="card-header alert alert-warning">
                <div>
                    Note: This editor is read-only. To edit settings or configuration, visit the appropriate page:
                </div>
                <ul className="mb-0">
                    <li>
                        <Link to="/site-admin/configuration">Site configuration</Link>
                    </li>
                    <li>
                        <Link to="/site-admin/external-services">External services</Link>
                    </li>
                    <li>
                        <Link to="/site-admin/global-settings">Global settings</Link>
                    </li>
                    <li>
                        <Link to="/site-admin/organizations">Organization settings</Link>
                    </li>
                    <li>
                        <Link to={`/users/${authenticatedUser.username}/settings`}>User settings</Link>
                    </li>
                </ul>
            </div>
            <DynamicallyImportedMonacoSettingsEditor
                value={allConfig ? JSON.stringify(allConfig, undefined, 2) : ''}
                jsonSchema={allConfigSchema}
                canEdit={false}
                height={800}
                isLightTheme={isLightTheme}
                history={history}
                readOnly={true}
            />
        </div>
    )
}

export const SiteAdminExportConfigPage = withAuthenticatedUser(AuthenticatedSiteAdminExportConfigPage)
