import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'

import { OrgSettingsAreaRoute } from '../../../org/settings/OrgSettingsArea'
import { orgSettingsAreaRoutes } from '../../../org/settings/routes'
import type { OrgExecutorSecretsListPageProps } from '../../executors/secrets/ExecutorSecretsListPage'

const OrgExecutorSecretsListPage = lazyComponent<OrgExecutorSecretsListPageProps, 'OrgExecutorSecretsListPage'>(
    () => import('../../executors/secrets/ExecutorSecretsListPage'),
    'OrgExecutorSecretsListPage'
)

export const enterpriseOrgSettingsAreaRoutes: readonly OrgSettingsAreaRoute[] = [
    ...orgSettingsAreaRoutes,
    {
        path: '/executors/secrets',
        render: props => <OrgExecutorSecretsListPage {...props} orgID={props.org.id} />,
        condition: ({ org: { viewerCanAdminister } }) => viewerCanAdminister,
    },
]
