import { Optional } from 'utility-types'

// eslint-disable-next-line no-restricted-imports
import { TourListState } from '@sourcegraph/web/src/tour/components/Tour/useTour'
import { MultiSelectState } from '@sourcegraph/wildcard'

import { BatchChangeState } from '../../graphql-operations'

import { SectionID, NoResultsSectionID } from './searchSidebar'

/**
 * Schema for temporary settings.
 */
export interface TemporarySettingsSchema {
    'search.collapsedSidebarSections': { [key in SectionID]?: boolean }
    'search.hiddenNoResultsSections': NoResultsSectionID[]
    'search.sidebar.revisions.tab': number
    'search.sidebar.collapsed': boolean // Used only on non-mobile sizes and when coreWorkflowImprovements.enabled is set
    'search.notepad.enabled': boolean
    'search.notepad.ctaSeen': boolean
    'search.notebooks.gettingStartedTabSeen': boolean
    'insights.freeGaAccepted': boolean
    'insights.freeGaExpiredAccepted': boolean
    'insights.wasMainPageOpen': boolean
    'npsSurvey.hasTemporarilyDismissed': boolean
    'npsSurvey.hasPermanentlyDismissed': boolean
    'user.lastDayActive': string | null
    'user.daysActiveCount': number
    'user.themePreference': string
    'signup.finishedWelcomeFlow': boolean
    'homepage.userInvites.tab': number
    'batches.defaultListFilters': MultiSelectState<BatchChangeState>
    'batches.downloadSpecModalDismissed': boolean
    'codeintel.badge.used': boolean
    'codeintel.referencePanel.redesign.ctaDismissed': boolean
    'codeintel.referencePanel.redesign.enabled': boolean
    'onboarding.quickStartTour': TourListState
    'characterKeyShortcuts.enabled': boolean
    'search.homepage.queryExamplesContent': {
        lastCachedTimestamp: string
        repositoryName: string
        filePath: string
        author: string
    }
    'search.results.collapseSmartSearch': boolean
    // TODO #41002: Remove this temporary setting.
    // This temporary setting is now turned on by default with no UI to toggle it off.
    'coreWorkflowImprovements.enabled_deprecated': boolean
}

/**
 * All temporary settings are possibly undefined. This is the actual schema that
 * should be used to force the consumer to check for undefined values.
 */
export type TemporarySettings = Optional<TemporarySettingsSchema>
