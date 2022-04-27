import { Optional } from 'utility-types'

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
    'search.onboarding.tourCancelled': boolean
    'search.contexts.ctaDismissed': boolean
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
    'signup.finishedWelcomeFlow': boolean
    'codemonitor.info.visible': boolean
    'homepage.userInvites.tab': number
    'integrations.vscode.lastDetectionTimestamp': number
    'integrations.jetbrains.lastDetectionTimestamp': number
    'cta.browserExtensionAlertDismissed': boolean
    'cta.ideExtensionAlertDismissed': boolean
    'batches.defaultListFilters': MultiSelectState<BatchChangeState>
    'batches.downloadSpecModalDismissed': boolean
    'codeintel.badge.used': boolean
}

/**
 * All temporary settings are possibly undefined. This is the actual schema that
 * should be used to force the consumer to check for undefined values.
 */
export type TemporarySettings = Optional<TemporarySettingsSchema>
