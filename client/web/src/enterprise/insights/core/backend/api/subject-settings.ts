import { from, Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { PlatformContext } from '@sourcegraph/shared/src/platform/context'

import { fetchLatestSubjectSettings } from '../requests/fetch-backend-insights'
import { SubjectSettingsResult } from '../types'

/**
 * Get settings of particular subject by id.
 * */
export const getSubjectSettings = (id: string): Observable<SubjectSettingsResult> =>
    fetchLatestSubjectSettings(id).pipe(
        map(settings => settings.settingsSubject?.latestSettings ?? { id: null, contents: '' })
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
