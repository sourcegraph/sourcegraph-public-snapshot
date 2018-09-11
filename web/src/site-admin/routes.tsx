import React from 'react'
import { SiteAdminAllUsersPage } from './SiteAdminAllUsersPage'
import { SiteAdminAnalyticsPage } from './SiteAdminAnalyticsPage'
import { SiteAdminAreaRoute } from './SiteAdminArea'
import { SiteAdminCodeIntelligencePage } from './SiteAdminCodeIntelligencePage'
import { SiteAdminConfigurationPage } from './SiteAdminConfigurationPage'
import { SiteAdminCreateUserPage } from './SiteAdminCreateUserPage'
import { SiteAdminOrgsPage } from './SiteAdminOrgsPage'
import { SiteAdminOverviewPage } from './SiteAdminOverviewPage'
import { SiteAdminPingsPage } from './SiteAdminPingsPage'
import { SiteAdminRepositoriesPage } from './SiteAdminRepositoriesPage'
import { SiteAdminSettingsPage } from './SiteAdminSettingsPage'
import { SiteAdminSurveyResponsesPage } from './SiteAdminSurveyResponsesPage'
import { SiteAdminTokensPage } from './SiteAdminTokensPage'
import { SiteAdminUpdatesPage } from './SiteAdminUpdatesPage'

export const siteAdminAreaRoutes: ReadonlyArray<SiteAdminAreaRoute> = [
    {
        // Render empty page if no page selected
        path: '',
        render: props => <SiteAdminOverviewPage {...props} />,
        exact: true,
    },
    {
        path: '/configuration',
        exact: true,
        render: props => <SiteAdminConfigurationPage {...props} />,
    },
    {
        path: '/global-settings',
        exact: true,
        render: props => <SiteAdminSettingsPage {...props} />,
    },
    {
        path: '/repositories',
        render: props => <SiteAdminRepositoriesPage {...props} />,
        exact: true,
    },
    {
        path: '/code-intelligence',
        render: props => <SiteAdminCodeIntelligencePage {...props} />,
        exact: true,
    },
    {
        path: '/organizations',
        render: props => <SiteAdminOrgsPage {...props} />,
        exact: true,
    },
    {
        path: '/users',
        exact: true,
        render: props => <SiteAdminAllUsersPage {...props} />,
    },
    {
        path: '/users/new',
        render: props => <SiteAdminCreateUserPage {...props} />,
        exact: true,
    },
    {
        path: '/tokens',
        exact: true,
        render: props => <SiteAdminTokensPage {...props} />,
    },
    {
        path: '/analytics',
        exact: true,
        render: props => <SiteAdminAnalyticsPage {...props} />,
    },
    {
        path: '/updates',
        render: props => <SiteAdminUpdatesPage {...props} />,
        exact: true,
    },
    {
        path: '/pings',
        render: props => <SiteAdminPingsPage {...props} />,
        exact: true,
    },
    {
        path: '/surveys',
        exact: true,
        render: props => <SiteAdminSurveyResponsesPage {...props} />,
    },
]
