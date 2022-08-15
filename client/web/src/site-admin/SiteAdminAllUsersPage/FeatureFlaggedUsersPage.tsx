import { withFeatureFlag } from '../../featureFlags/withFeatureFlag'

import { UsersManagement } from './UserManagement'

import { SiteAdminAllUsersPage } from '.'

export const FeatureFlaggedUsersPage = withFeatureFlag('user-management', UsersManagement, SiteAdminAllUsersPage)
