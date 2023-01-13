import { StreamingSearchResultsListProps } from '@sourcegraph/search-ui'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { SearchContextProps } from '@sourcegraph/shared/src/search'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'

import { AuthenticatedUser } from '../../auth'
import { BatchChangesProps } from '../../batches'
import { CodeIntelligenceProps } from '../../codeintel'
import { BreadcrumbSetters } from '../../components/Breadcrumbs'
import { ActionItemsBarProps } from '../../extensions/components/ActionItemsBar'
import { ExternalLinkFields } from '../../graphql-operations'
import { CodeInsightsProps } from '../../insights/types'
import { SearchStreamingProps } from '../../search'
import { RouteDescriptor } from '../../util/contributions'
import { HoverThresholdProps } from '../RepoContainer'
import { RepoHeaderContributionsLifecycleProps } from '../RepoHeader'

import { RepoSettingsAreaRoute } from './RepoSettingsArea'
import { RepoSettingsSideBarGroup } from './RepoSettingsSidebar'

export interface RepoSettingsContainerRoute extends RouteDescriptor<RepoSettingsContainerContext> {}

export interface RepoSettingsContainerContext
    extends RepoHeaderContributionsLifecycleProps,
        SettingsCascadeProps,
        ExtensionsControllerProps,
        PlatformContextProps,
        ThemeProps,
        HoverThresholdProps,
        TelemetryProps,
        Pick<SearchContextProps, 'selectedSearchContextSpec' | 'searchContextsEnabled'>,
        BreadcrumbSetters,
        ActionItemsBarProps,
        SearchStreamingProps,
        Pick<StreamingSearchResultsListProps, 'fetchHighlightedFileLineRanges'>,
        CodeIntelligenceProps,
        BatchChangesProps,
        CodeInsightsProps {
    // repo: RepositoryFields | undefined
    // repo: RepositoryFields
    repoName: string
    // resolvedRevisionOrError: ResolvedRevision | ErrorLike | undefined
    authenticatedUser: AuthenticatedUser | null
    repoSettingsAreaRoutes: readonly RepoSettingsAreaRoute[]
    repoSettingsSidebarGroups: readonly RepoSettingsSideBarGroup[]

    /** The URL route match for {@link RepoContainer}. */
    routePrefix: string

    onDidUpdateExternalLinks: (externalLinks: ExternalLinkFields[] | undefined) => void

    globbing: boolean

    isMacPlatform: boolean

    isSourcegraphDotCom: boolean
}
