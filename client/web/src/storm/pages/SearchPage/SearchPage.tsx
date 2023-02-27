import { FC } from 'react'

import { useLegacyRouteContext } from '../../../LegacyRouteContext'

import { usePreloadedQueryData } from './SearchPage.loader'
import { SearchPageContent, getShouldShowAddCodeHostWidget } from './SearchPageContent'

export const SearchPage: FC = () => {
    const { data } = usePreloadedQueryData()
    const { authenticatedUser } = useLegacyRouteContext()

    const shouldShowAddCodeHostWidget = getShouldShowAddCodeHostWidget({
        isAddCodeHostWidgetEnabled: !!data?.evaluateFeatureFlag,
        isSiteAdmin: authenticatedUser?.siteAdmin,
        externalServicesCount: data?.externalServices.totalCount,
    })

    return <SearchPageContent shouldShowAddCodeHostWidget={shouldShowAddCodeHostWidget} />
}
