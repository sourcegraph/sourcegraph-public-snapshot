import { ExtensionsProps as GenericExtensionsProps } from '@sourcegraph/extensions-client-common/lib/context'
import { Controller as ExtensionsContextController } from '@sourcegraph/extensions-client-common/lib/controller'
import { ConfigurationSubject, gqlToCascade } from '@sourcegraph/extensions-client-common/lib/settings'
import Loader from '@sourcegraph/icons/lib/Loader'
import Warning from '@sourcegraph/icons/lib/Warning'
import { isEqual } from 'lodash'
import { concat, Observable } from 'rxjs'
import { distinctUntilChanged, map, switchMap, take } from 'rxjs/operators'
import { gql, queryGraphQL } from '../backend/graphql'
import * as GQL from '../backend/graphqlschema'
import { editConfiguration } from '../configuration/backend'
import { configurationCascade, toGQLKeyPath } from '../settings/configuration'
import { refreshConfiguration } from '../user/settings/backend'

export interface ExtensionsProps extends GenericExtensionsProps<ConfigurationSubject> {}

export function createExtensionsContextController(): ExtensionsContextController<ConfigurationSubject> {
    return new ExtensionsContextController<ConfigurationSubject>({
        configurationCascade: configurationCascade.pipe(
            map(gqlToCascade),
            distinctUntilChanged((a, b) => isEqual(a, b))
        ),
        updateExtensionSettings,
        queryGraphQL: (request, variables) =>
            queryGraphQL(
                gql`
                    ${request}
                `,
                variables
            ),
        icons: {
            Loader: Loader as React.ComponentType<{ className: 'icon-inline' }>,
            Warning: Warning as React.ComponentType<{ className: 'icon-inline' }>,
        },
    })
}

function updateExtensionSettings(
    subject: GQL.ConfigurationSubject | GQL.IConfigurationSubject | { id: GQL.ID },
    args: {
        extensionID: string
        edit?: GQL.IConfigurationEdit
        enabled?: boolean
        remove?: boolean
    }
): Observable<void> {
    return configurationCascade.pipe(
        take(1),
        switchMap(configurationCascade => {
            const subjectConfig = configurationCascade.subjects.find(s => s.id === subject.id)
            if (!subjectConfig) {
                throw new Error(`no configuration subject: ${subject.id}`)
            }
            const lastID = subjectConfig.latestSettings ? subjectConfig.latestSettings.id : null

            let edit: GQL.IConfigurationEdit
            if (args.edit) {
                edit = args.edit
            } else if (typeof args.enabled === 'boolean') {
                edit = { keyPath: toGQLKeyPath(['extensions', args.extensionID]), value: args.enabled }
            } else if (args.remove) {
                edit = { keyPath: toGQLKeyPath(['extensions', args.extensionID]), value: null }
            } else {
                throw new Error('no edit')
            }
            return editConfiguration(subject.id, lastID, edit)
        }),
        switchMap(() => concat(refreshConfiguration(), [void 0]))
    )
}

export function updateUserExtensionSettings(args: {
    extensionID: string
    enabled?: boolean
    edit?: GQL.IConfigurationEdit
}): Observable<void> {
    return configurationCascade.pipe(
        take(1),
        switchMap(configurationCascade => {
            // Only support configuring extension settings in user settings with this action.
            const subject = configurationCascade.subjects[configurationCascade.subjects.length - 1]
            return updateExtensionSettings(subject, args)
        })
    )
}
