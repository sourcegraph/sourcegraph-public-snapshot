import { useEffect, useMemo } from 'react'

import * as H from 'history'

import { UserHistory } from '@sourcegraph/shared/src/components/UserHistory'

import { useFeatureFlag } from '../featureFlags/useFeatureFlag'
import { parseBrowserRepoURL } from '../util/url'

export function useUserHistory(history: H.History, isRepositoryRelatedPage: boolean): UserHistory | undefined {
    const { location } = history
    const [isFeatureEnabled, status] = useFeatureFlag('user-history-ranking', true)
    const isEnabled = status === 'loaded' && isFeatureEnabled
    const userHistory = useMemo(() => new UserHistory(parseBrowserRepoURL), [])
    useEffect(() => {
        if (isEnabled && isRepositoryRelatedPage) {
            userHistory.onLocation(location)
        }
    }, [isEnabled, userHistory, location, isRepositoryRelatedPage])
    return isEnabled ? userHistory : undefined
}
