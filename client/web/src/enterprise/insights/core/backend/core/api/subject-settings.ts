import { from, Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { dataOrThrowErrors, gql } from '@sourcegraph/http-client'
import { PlatformContext } from '@sourcegraph/shared/src/platform/context'

import { requestGraphQL } from '../../../../../../backend/graphql'
import { SubjectSettings, SubjectSettingsResult, SubjectSettingsVariables } from '../../../../../../graphql-operations'

const SUBJECT_SETTINGS_FRAGMENT = gql`
    fragment SubjectSettings on Settings {
        id
        contents
    }
`

/**
 * Returns settings of a particular subject by its id.
 */
export const getSubjectSettings = (id: string): Observable<SubjectSettings> =>
    requestGraphQL<SubjectSettingsResult, SubjectSettingsVariables>(
        gql`
            query SubjectSettings($id: ID!) {
                settingsSubject(id: $id) {
                    latestSettings {
                        ...SubjectSettings
                    }
                }
            }
            ${SUBJECT_SETTINGS_FRAGMENT}
        `,
        { id }
    ).pipe(
        map(dataOrThrowErrors),
        map(settings => settings.settingsSubject?.latestSettings ?? { id: 0, contents: '' })
    )

/**
 * Update settings content of the subject by its id.
 *
 * @param context Context of host app. Used to trigger setting update.
 * @param subjectId Id of the subject
 * @param content New content of the settings
 * */
export const updateSubjectSettings = (
    context: Pick<PlatformContext, 'updateSettings'>,
    subjectId: string,
    content: string
): Observable<void> => from(context.updateSettings(subjectId, content))
