import type { FC } from 'react'

import { gql, useQuery } from '@sourcegraph/http-client'
import { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'

import type { AuthenticatedUser } from '../../auth'
import { useFeatureFlag } from '../../featureFlags/useFeatureFlag'
import type { ExternalServicesTotalCountResult } from '../../graphql-operations'
import { SearchPageContent, getShouldShowAddCodeHostWidget } from '../../storm/pages/SearchPage/SearchPageContent'

export interface SearchPageProps extends TelemetryV2Props {
    authenticatedUser: AuthenticatedUser | null
}

export const SearchPage: FC<SearchPageProps> = props => {
    const shouldShowAddCodeHostWidget = useShouldShowAddCodeHostWidget(props.authenticatedUser)
    return (
        <SearchPageContent
            shouldShowAddCodeHostWidget={shouldShowAddCodeHostWidget}
            telemetryRecorder={props.telemetryRecorder}
        />
    )
}

const EXTERNAL_SERVICES_TOTAL_COUNT = gql`
    query ExternalServicesTotalCount {
        externalServices {
            totalCount
        }
    }
`

function useShouldShowAddCodeHostWidget(authenticatedUser: AuthenticatedUser | null): boolean | undefined {
    const [isAddCodeHostWidgetEnabled] = useFeatureFlag('plg-enable-add-codehost-widget', false)
    const { data } = useQuery<ExternalServicesTotalCountResult>(EXTERNAL_SERVICES_TOTAL_COUNT, {})

    return getShouldShowAddCodeHostWidget({
        isAddCodeHostWidgetEnabled,
        isSiteAdmin: authenticatedUser?.siteAdmin,
        externalServicesCount: data?.externalServices.totalCount,
    })
}
