import { RouteComponentProps } from 'react-router'
import { fetchAllConfigAndSettings } from './backend'
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

/**
 * Minimal shape of a JSON Schema. These values are treated as opaque, so more specific types are
 * not needed.
 */
interface JSONSchema {
    $id: string
    definitions?: Record<string, { type: string }>
}

const externalServices: [ExternalServiceKind, JSONSchema][] = [
    [ExternalServiceKind.AWSCODECOMMIT, awsCodeCommitJSON],
    [ExternalServiceKind.BITBUCKETCLOUD, bitbucketCloudSchemaJSON],
    [ExternalServiceKind.BITBUCKETSERVER, bitbucketServerSchemaJSON],
    [ExternalServiceKind.GITHUB, githubSchemaJSON],
    [ExternalServiceKind.GITLAB, gitlabSchemaJSON],
    [ExternalServiceKind.GITOLITE, gitoliteSchemaJSON],
    [ExternalServiceKind.OTHER, otherExternalServiceSchemaJSON],
    [ExternalServiceKind.PHABRICATOR, phabricatorSchemaJSON],
]

const allConfigSchema = {
    $id: 'all.schema.json#',
    allowComments: true,
    additionalProperties: false,
    properties: {
        site: siteSchemaJSON,
        externalServices: {
            type: 'object',
            properties: externalServices.reduce<
                Partial<{ [k in ExternalServiceKind]: { type: string; items: JSONSchema } }>
            >((properties, [kind, schema]) => {
                properties[kind] = {
                    type: 'array',
                    items: schema,
                }
                return properties
            }, {}),
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
    definitions: externalServices
        .map(([, schema]) => schema.definitions)
        .concat([siteSchemaJSON.definitions, settingsSchemaJSON.definitions])
        .reduce(
            (definitions, theseDefinitions) =>
                theseDefinitions
                    ? {
                          ...definitions,
                          ...theseDefinitions,
                      }
                    : definitions,
            {}
        ),
}

interface Props extends RouteComponentProps {
    isLightTheme: boolean
}

export const SiteAdminExportCeonfigPage: React.FunctionComponent<Props> = ({ isLightTheme, history }) => {
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
                Note: This editor is for export purposes only. You may edit the contents and use auto-complete, but
                changes will not be saved. Reloading the page will erase any of the changes you make in this editor.
            </div>
            <DynamicallyImportedMonacoSettingsEditor
                value={allConfig ? JSON.stringify(allConfig, undefined, 2) : ''}
                jsonSchema={allConfigSchema}
                canEdit={false}
                height={800}
                isLightTheme={isLightTheme}
                history={history}
            />
        </div>
    )
}
