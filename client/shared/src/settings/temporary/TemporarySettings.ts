import type { Optional } from 'utility-types'

import type { BatchChangeState } from '../../graphql-operations'

import type { DiffMode } from './diffMode'
import type { RecentSearch } from './recentSearches'
import type { SectionID, NoResultsSectionID } from './searchSidebar'
import type { TourListState } from './tourState'

// Prior to this type we store in settings list of MultiSelectState
// we no longer use MultiSelect UI but for backward compatibility we still
// have to store and parse the old version of batch changes filters
export interface LegacyBatchChangesFilter {
    label: string
    value: BatchChangeState
}

export interface UserOnboardingConfig {
    skipped: boolean
    userinfo?: {
        repo: string
        email: string
        language: string
    }
}

/**
 * Schema for temporary settings.
 */
export interface TemporarySettingsSchema {
    'search.collapsedSidebarSections': { [key in SectionID]?: boolean }
    'search.hiddenNoResultsSections': NoResultsSectionID[]
    'search.sidebar.revisions.tab': number
    'search.sidebar.collapsed': boolean // Used only on non-mobile sizes and when coreWorkflowImprovements.enabled is set
    'search.notebooks.gettingStartedTabSeen': boolean
    'insights.freeGaAccepted': boolean
    'insights.freeGaExpiredAccepted': boolean
    'insights.wasMainPageOpen': boolean
    'insights.lastVisitedDashboardId': string | null
    'npsSurvey.hasTemporarilyDismissed': boolean
    'npsSurvey.hasPermanentlyDismissed': boolean
    'user.lastDayActive': string | null
    'user.daysActiveCount': number
    'user.themePreference': string
    'signup.finishedWelcomeFlow': boolean
    'homepage.userInvites.tab': number
    'batches.defaultListFilters': LegacyBatchChangesFilter[]
    'batches.downloadSpecModalDismissed': boolean
    'codeintel.badge.used': boolean
    'codeintel.referencePanel.redesign.ctaDismissed': boolean
    'onboarding.quickStartTour': TourListState
    'onboarding.userconfig': UserOnboardingConfig
    'characterKeyShortcuts.enabled': boolean
    'search.homepage.queryExamplesContent': {
        lastCachedTimestamp: string
        repositoryName: string
        filePath: string
    }
    'search.results.collapseSmartSearch': boolean
    'search.results.collapseUnownedResultsAlert': boolean
    'search.input.recentSearches': RecentSearch[]
    /**
     * Keeps track of which of the query examples shown as suggestions
     * the user has used so that we don't suggest them anymore.
     */
    'search.input.usedExamples': string[]
    'search.input.usedInlineHistory': boolean

    /** Let users quickly switch between the v1 and v2 query inputs. */
    'search.input.experimental': boolean

    'batches.minSavedPerChangeset': number
    'search.notebooks.minSavedPerView': number
    'repo.commitPage.diffMode': DiffMode
    'setup.activeStepId': string
    'app-setup.activeStepId': string
    'own.panelExplanationHidden': boolean
    'cody.showSidebar': boolean
    'cody.blobPageCta.dismissed': boolean
    'cody.searchPageCta.dismissed': boolean
    'cody.chatPageCta.dismissed': boolean
    'cody.survey.submitted': boolean
    'app.codyStandalonePage.selectedRepo': string
    'cody.contextCallout.dismissed': boolean
    'admin.hasDismissedCodeHostPrivacyWarning': boolean
    'admin.hasCompletedLicenseCheck': boolean
    'simple.search.toggle': boolean
    'cody.onboarding.completed': boolean
    'cody.onboarding.step': number

    /** OpenCodeGraph */
    'openCodeGraph.annotations.visible': boolean
}

/**
 * All temporary settings are possibly undefined. This is the actual schema that
 * should be used to force the consumer to check for undefined values.
 */
export type TemporarySettings = Optional<TemporarySettingsSchema>

// TypeScript doesn't have a concept of "exhaustive" list or sets, so we use
// a record instead.
const TEMPORARY_SETTINGS: Record<keyof TemporarySettings, null> = {
    'search.collapsedSidebarSections': null,
    'search.hiddenNoResultsSections': null,
    'search.sidebar.revisions.tab': null,
    'search.sidebar.collapsed': null,
    'search.notebooks.gettingStartedTabSeen': null,
    'insights.freeGaAccepted': null,
    'insights.freeGaExpiredAccepted': null,
    'insights.wasMainPageOpen': null,
    'insights.lastVisitedDashboardId': null,
    'npsSurvey.hasTemporarilyDismissed': null,
    'npsSurvey.hasPermanentlyDismissed': null,
    'user.lastDayActive': null,
    'user.daysActiveCount': null,
    'user.themePreference': null,
    'signup.finishedWelcomeFlow': null,
    'homepage.userInvites.tab': null,
    'batches.defaultListFilters': null,
    'batches.downloadSpecModalDismissed': null,
    'codeintel.badge.used': null,
    'codeintel.referencePanel.redesign.ctaDismissed': null,
    'onboarding.quickStartTour': null,
    'onboarding.userconfig': null,
    'characterKeyShortcuts.enabled': null,
    'search.homepage.queryExamplesContent': null,
    'search.results.collapseSmartSearch': null,
    'search.results.collapseUnownedResultsAlert': null,
    'search.input.recentSearches': null,
    /**
     * Keeps track of which of the query examples shown as suggestions
     * the user has used so that we don't suggest them anymore.
     */
    'search.input.usedExamples': null,
    'search.input.usedInlineHistory': null,
    'search.input.experimental': null,
    'batches.minSavedPerChangeset': null,
    'search.notebooks.minSavedPerView': null,
    'repo.commitPage.diffMode': null,
    'setup.activeStepId': null,
    'app-setup.activeStepId': null,
    'own.panelExplanationHidden': null,
    'cody.showSidebar': null,
    'cody.blobPageCta.dismissed': null,
    'cody.searchPageCta.dismissed': null,
    'cody.chatPageCta.dismissed': null,
    'cody.survey.submitted': null,
    'app.codyStandalonePage.selectedRepo': null,
    'cody.contextCallout.dismissed': null,
    'admin.hasDismissedCodeHostPrivacyWarning': null,
    'admin.hasCompletedLicenseCheck': null,
    'simple.search.toggle': null,
    'cody.onboarding.completed': null,
    'cody.onboarding.step': null,
    'openCodeGraph.annotations.visible': null,
}

export const TEMPORARY_SETTINGS_KEYS = Object.keys(TEMPORARY_SETTINGS) as readonly (keyof TemporarySettings)[]
