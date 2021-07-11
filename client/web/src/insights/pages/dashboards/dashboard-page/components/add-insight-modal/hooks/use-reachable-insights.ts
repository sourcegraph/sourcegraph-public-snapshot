import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings';
import { isErrorLike } from '@sourcegraph/shared/src/util/errors';

import { Settings } from '../../../../../../../schema/settings.schema';
import { Insight, InsightConfiguration, isInsightSettingKey } from '../../../../../../core/types';
import { createInsightFromSettings } from '../../../../../../hooks/use-insight/use-insight';

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
export function useReachableInsights(props: UseReachableInsightsProps): Insight[] {
    const { settingsCascade, ownerId } = props

    if (!settingsCascade.subjects) {
        return []
    }

    const configureSubject = settingsCascade.subjects
        .find(configureSubject => configureSubject.subject.id === ownerId)

    const ownerSubject = configureSubject?.subject
    const ownerSubjectSettings = configureSubject?.settings

    if (!ownerSubject) {
        return []
    }

    switch (ownerSubject.__typename) {
        case 'User': {
            const final = settingsCascade.final

            if (!final ||isErrorLike(final)) {
                return []
            }

            return Object.keys(final)
                .filter(isInsightSettingKey)
                .map(key => createInsightFromSettings({
                    insightKey: key,
                    ownerId: ownerSubject.id,
                    insightConfiguration: final[key] as InsightConfiguration
                }) )
        }

        case 'Org': {
            if (!ownerSubjectSettings || isErrorLike(ownerSubjectSettings)) {
                return []
            }

            return Object.keys(ownerSubjectSettings)
                .filter(isInsightSettingKey)
                .map(key => createInsightFromSettings({
                    insightKey: key,
                    ownerId: ownerSubject.id,
                    insightConfiguration: ownerSubjectSettings[key] as InsightConfiguration
                }))
        }

        default: return []
    }
}
