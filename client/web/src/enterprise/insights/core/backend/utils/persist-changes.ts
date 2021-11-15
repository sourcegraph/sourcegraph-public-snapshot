import { groupBy } from 'lodash'
import { forkJoin, Observable } from 'rxjs'
import { switchMap } from 'rxjs/operators'

import { PlatformContext } from '@sourcegraph/shared/src/platform/context'

import { applyEditOperations, SettingsOperation } from '../../settings-action/edits'
import { getSubjectSettings, updateSubjectSettings } from '../api/subject-settings'

export const persistChanges = (
    platformContext: Pick<PlatformContext, 'updateSettings'>,
    operations: SettingsOperation[]
): Observable<void[]> => {
    const subjectsToUpdate = groupBy(operations, operation => operation.subjectId)

    const subjectUpdateRequests = Object.keys(subjectsToUpdate).map(subjectId => {
        const editOperations = subjectsToUpdate[subjectId]

        return getSubjectSettings(subjectId).pipe(
            switchMap(settings => {
                // Modify this jsonc file according to this subject's operations
                const nextSubjectSettings = applyEditOperations(settings.contents, editOperations)

                return updateSubjectSettings(platformContext, subjectId, nextSubjectSettings)
            })
        )
    })

    return forkJoin(subjectUpdateRequests)
}
