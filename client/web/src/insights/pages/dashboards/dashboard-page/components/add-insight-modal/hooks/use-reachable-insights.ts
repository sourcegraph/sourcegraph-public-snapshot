import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { isErrorLike } from '@sourcegraph/shared/src/util/errors'

import { Settings } from '../../../../../../../schema/settings.schema'
import { Insight, InsightConfiguration, isInsightSettingKey } from '../../../../../../core/types'
import {
    isSubjectInsightSupported,
    SUBJECT_SHARING_LEVELS,
    SupportedInsightSubject,
} from '../../../../../../core/types/subjects'
import { getDashboardOwnerInfo } from '../../../../../../hooks/use-dashboards/utils'
import { createInsightFromSettings } from '../../../../../../hooks/use-insight/use-insight'

export interface UseReachableInsightsProps extends SettingsCascadeProps<Settings> {
    /**
     * Subject dashboard owner id.
     */
    ownerId: string
}

export type ReachableInsight = Insight & {
    owner: {
        id: string
        name: string
    }
}

/**
 * Returns all reachable subject's insights by owner id.
 *
 * User subject has access to all insights from all organizations and global site settings.
 * Organization subject has access to only its insights.
 */
export function useReachableInsights(props: UseReachableInsightsProps): ReachableInsight[] {
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

            return Object.keys(settings)
                .filter(isInsightSettingKey)
                .map(key => {
                    const insight = createInsightFromSettings({
                        insightKey: key,
                        ownerId: subject.id,
                        insightConfiguration: settings[key] as InsightConfiguration,
                    })

                    return {
                        ...insight,
                        // Extend common insight object with owner info
                        owner: subjectOwnerInfo,
                    }
                })
        })
}
