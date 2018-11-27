import { from } from 'rxjs'
import { map, switchMap, take } from 'rxjs/operators'
import { SettingsEdit } from '../api/protocol'
import { dataOrThrowErrors, gql, graphQLContent } from '../graphql/graphql'
import * as GQL from '../graphql/schema'
import { PlatformContext } from '../platform/context'
import { isErrorLike } from '../util/errors'

export function updateSettings(
    { settingsCascade, queryGraphQL }: Pick<PlatformContext, 'settingsCascade' | 'queryGraphQL'>,
    subject: GQL.ID,
    args: SettingsEdit,
    applySettingsEdit: (
        { queryGraphQL }: Pick<PlatformContext, 'queryGraphQL'>,
        subject: GQL.ID,
        lastID: number | null,
        edit: GQL.ISettingsEdit
    ) => Promise<void>
): Promise<void> {
    return from(settingsCascade)
        .pipe(
            take(1),
            switchMap(settingsCascade => {
                if (!settingsCascade.subjects) {
                    throw new Error('settings not available')
                }
                if (isErrorLike(settingsCascade.subjects)) {
                    throw new Error(`settings not available due to error: ${settingsCascade.subjects.message}`)
                }
                const subjectSettings = settingsCascade.subjects.find(s => s.subject.id === subject)
                if (!subjectSettings) {
                    throw new Error(`no settings subject: ${subject}`)
                }
                if (isErrorLike(subjectSettings.settings)) {
                    throw new Error(`settings subject error: ${subjectSettings.settings.message}`)
                }
                const lastID = subjectSettings.settings ? subjectSettings.lastID : null

                return applySettingsEdit({ queryGraphQL }, subject, lastID, {
                    keyPath: toGQLKeyPath(args.path),
                    value: args.value,
                })
            })
        )
        .toPromise()
}

function toGQLKeyPath(keyPath: (string | number)[]): GQL.IKeyPathSegment[] {
    return keyPath.map(v => (typeof v === 'string' ? { property: v } : { index: v }))
}

// NOTE: uses configurationMutation (not settingsMutation) and editConfiguration (not editSettings) for backcompat.
export function mutateSettings(
    { queryGraphQL }: Pick<PlatformContext, 'queryGraphQL'>,
    subject: GQL.ID,
    lastID: number | null,
    edit: GQL.IConfigurationEdit
): Promise<void> {
    return from(
        queryGraphQL(
            gql`
                mutation EditConfiguration($subject: ID!, $lastID: Int, $edit: ConfigurationEdit!) {
                    configurationMutation(input: { subject: $subject, lastID: $lastID }) {
                        editConfiguration(edit: $edit) {
                            empty {
                                alwaysNil
                            }
                        }
                    }
                }
            `[graphQLContent],
            { subject, lastID, edit }
        )
    )
        .pipe(
            map(dataOrThrowErrors),
            map(() => undefined)
        )
        .toPromise()
}
