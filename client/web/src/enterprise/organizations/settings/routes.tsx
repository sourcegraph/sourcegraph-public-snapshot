import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'

import { OrgSettingsAreaRoute } from '../../../org/settings/OrgSettingsArea'
import { orgSettingsAreaRoutes } from '../../../org/settings/routes'
import type { ExecutorsOrgAreaProps } from '../../executors/ExecutorsOrgArea'

const ExecutorsOrgArea = lazyComponent<ExecutorsOrgAreaProps, 'ExecutorsOrgArea'>(
    () => import('../../executors/ExecutorsOrgArea'),
    'ExecutorsOrgArea'
)

export const enterpriseOrgSettingsAreaRoutes: readonly OrgSettingsAreaRoute[] = [
    ...orgSettingsAreaRoutes,
    {
        path: '/executors',
        render: props => <ExecutorsOrgArea {...props} namespaceID={props.org.id} />,
        condition: ({ org: { viewerCanAdminister } }) => viewerCanAdminister,
    },
]
