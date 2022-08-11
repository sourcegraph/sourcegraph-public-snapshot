import { withFeatureFlag } from '../../featureFlags/withFeatureFlag'

import { UsersManagement } from './New'

import { SiteAdminAllUsersPage } from '.'

export const FeatureFlaggedUsersPage = withFeatureFlag('user-management', UsersManagement, SiteAdminAllUsersPage)
