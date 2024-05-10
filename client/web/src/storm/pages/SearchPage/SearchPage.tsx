import type { FC } from 'react'

import { useLegacyContext_onlyInStormRoutes } from '../../../LegacyRouteContext'

import { usePreloadedQueryData } from './SearchPage.loader'
import { SearchPageContent, getShouldShowAddCodeHostWidget } from './SearchPageContent'

export const SearchPage: FC = () => {
    const { data } = usePreloadedQueryData()
    const { authenticatedUser, isSourcegraphDotCom } = useLegacyContext_onlyInStormRoutes()

    const shouldShowAddCodeHostWidget = getShouldShowAddCodeHostWidget({
        isAddCodeHostWidgetEnabled: !!data?.codehostWidgetFlag,
        isSiteAdmin: authenticatedUser?.siteAdmin,
        externalServicesCount: data?.externalServices.totalCount,
    })

    return (
        <SearchPageContent
            shouldShowAddCodeHostWidget={shouldShowAddCodeHostWidget || true}
            isSourcegraphDotCom={isSourcegraphDotCom}
        />
    )
}
