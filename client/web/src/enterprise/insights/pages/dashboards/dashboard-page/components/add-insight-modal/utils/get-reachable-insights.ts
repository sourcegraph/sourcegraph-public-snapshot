import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { isErrorLike } from '@sourcegraph/shared/src/util/errors'
import { isDefined } from '@sourcegraph/shared/src/util/types'

import { Settings } from '../../../../../../../../schema/settings.schema'
import { ReachableInsight } from '../../../../../../core/backend/code-insights-backend-types'
import { INSIGHTS_ALL_REPOS_SETTINGS_KEY } from '../../../../../../core/types'
import {
    isSubjectInsightSupported,
    SUBJECT_SHARING_LEVELS,
    SupportedInsightSubject,
} from '../../../../../../core/types/subjects'
import { getDashboardOwnerInfo } from '../../../../../../hooks/use-dashboards/utils'
import { parseInsightFromSubject } from '../../../../../../hooks/use-insight/use-insight'

export interface UseReachableInsightsProps extends SettingsCascadeProps<Settings> {
    /**
     * Subject dashboard owner id.
     */
    ownerId: string
}

/**
 * Returns all reachable subject's insights by owner id.
 *
 * User subject has access to all insights from all organizations and global site settings.
 * Organization subject has access to only its insights.
 */
export function getReachableInsights(props: UseReachableInsightsProps): ReachableInsight[] {
    const { settingsCascade, ownerId } = props

    if (!settingsCascade.subjects) {
        return []
    }

    const ownerConfigureSubject = settingsCascade.subjects.find(
        configureSubject => configureSubject.subject.id === ownerId
    )

    if (!ownerConfigureSubject) {
        return []
    }

    const subjectSharingLevel = SUBJECT_SHARING_LEVELS[ownerConfigureSubject.subject.__typename]
    const availableSubjects = settingsCascade.subjects.filter(
        configuredSubject => SUBJECT_SHARING_LEVELS[configuredSubject.subject.__typename] > subjectSharingLevel
    )

    const subjectsWithReachableInsights = [ownerConfigureSubject, ...availableSubjects]

    return subjectsWithReachableInsights
        .filter(configureSubject => isSubjectInsightSupported(configureSubject.subject))
        .flatMap(configureSubject => {
            const { settings, subject } = configureSubject

            if (!settings || isErrorLike(settings)) {
                return []
            }

            const subjectOwnerInfo = getDashboardOwnerInfo(subject as SupportedInsightSubject)
            const possibleInsightKeys = [
                ...Object.keys(settings),
                ...Object.keys(settings?.[INSIGHTS_ALL_REPOS_SETTINGS_KEY] ?? {}),
            ]

            return Object.keys(possibleInsightKeys)
                .map(key => parseInsightFromSubject(key, configureSubject))
                .filter(isDefined)
                .map(insight => ({
                    ...insight,
                    // Extend common insight object with owner info
                    owner: subjectOwnerInfo,
                }))
        })
}
