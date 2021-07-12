import { useMemo } from 'react'

import { SettingsCascadeOrError, SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { isErrorLike } from '@sourcegraph/shared/src/util/errors'

import { Settings } from '../../../schema/settings.schema'
import { Insight, InsightConfiguration } from '../../core/types'

export interface UseInsightProps extends SettingsCascadeProps<Settings> {
    insightId: string
}

/**
 * Returns insight from the setting cascade.
 */
export function useInsight(props: UseInsightProps): Insight | null {
    const { settingsCascade, insightId } = props

    return useMemo(() => findInsightById(settingsCascade, insightId), [settingsCascade, insightId])
}

export function findInsightById(settingsCascade: SettingsCascadeOrError<Settings>, insightId: string): Insight | null {
    const subjects = settingsCascade.subjects

    const subject = subjects?.find(({ settings }) => settings && !isErrorLike(settings) && !!settings[insightId])

    if (!subject?.settings || isErrorLike(subject.settings)) {
        return null
    }

    const insightConfiguration = subject.settings[insightId] as InsightConfiguration

    return {
        id: insightId,
        visibility: subject.subject.id,
        ...insightConfiguration,
    }
}

interface InsightsInputs {
    insightKey: string
    insightConfiguration: InsightConfiguration
    ownerId: string
}

export function createInsightFromSettings(input: InsightsInputs): Insight {
    const { insightKey, ownerId, insightConfiguration } = input

    return {
        id: insightKey,
        visibility: ownerId,
        ...insightConfiguration,
    }
}
