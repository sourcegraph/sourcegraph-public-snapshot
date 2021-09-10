import { useCallback, useState } from 'react'

import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { ErrorLike } from '@sourcegraph/shared/src/util/errors'

import { Insight, InsightTypePrefix } from '../../core/types'
import { usePersistEditOperations } from '../use-persist-edit-operations'

import { getDeleteInsightEditOperations } from './delete-helpers'

export interface UseDeleteInsightProps extends SettingsCascadeProps, PlatformContextProps<'updateSettings'> {}

export interface UseDeleteInsightAPI {
    delete: (insight: Pick<Insight, 'id' | 'title'>) => Promise<void>
    loading: boolean
    error: ErrorLike | undefined
}

/**
 * Returns delete handler that deletes insight from all subject settings and from all dashboards
 * that include this insight.
 */
export function useDeleteInsight(props: UseDeleteInsightProps): UseDeleteInsightAPI {
    const { settingsCascade, platformContext } = props
    const { persist } = usePersistEditOperations({ platformContext })

    const [loading, setLoading] = useState<boolean>(false)
    const [error, setError] = useState<ErrorLike | undefined>()

    const handleDelete = useCallback(
        async (insight: Pick<Insight, 'id' | 'title'>) => {
            const shouldDelete = window.confirm(`Are you sure you want to delete the insight "${insight.title}"?`)

            // Prevent double call if we already have ongoing request
            if (loading || !shouldDelete) {
                return
            }

            setLoading(true)
            setError(undefined)

            // For backward compatibility with old code stats insight api we have to delete
            // this insight in a special way. See link below for more information.
            // https://github.com/sourcegraph/sourcegraph-code-stats-insights/blob/master/src/code-stats-insights.ts#L33
            const isOldCodeStatsInsight = insight.id === `${InsightTypePrefix.langStats}.language`

            const keyForSearchInSettings = isOldCodeStatsInsight
                ? // Hardcoded value of id from old version of stats insight extension API
                  'codeStatsInsights.query'
                : insight.id

            try {
                const deleteInsightOperations = getDeleteInsightEditOperations({
                    insightId: keyForSearchInSettings,
                    settingsCascade,
                })

                await persist(deleteInsightOperations)
            } catch (error) {
                // TODO [VK] Improve error UI for deleting
                console.error(error)
                setError(error)
            }

            setLoading(false)
        },
        [persist, settingsCascade, loading]
    )

    return { delete: handleDelete, loading, error }
}
