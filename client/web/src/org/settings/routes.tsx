import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'
import { Text } from '@sourcegraph/wildcard'

import { SiteAdminAlert } from '../../site-admin/SiteAdminAlert'

import { DeleteOrg } from './DeleteOrg'
import { OrgSettingsAreaRoute } from './OrgSettingsArea'

const SettingsArea = lazyComponent(() => import('../../settings/SettingsArea'), 'SettingsArea')

export const orgSettingsAreaRoutes: readonly OrgSettingsAreaRoute[] = [
    {
        path: '',
        exact: true,
        render: props => (
            <div>
                <SettingsArea
                    {...props}
                    subject={props.org}
                    extraHeader={
                        <>
                            {props.authenticatedUser && props.org.viewerCanAdminister && !props.org.viewerIsMember && (
                                <SiteAdminAlert className="sidebar__alert">
                                    Viewing settings for <strong>{props.org.name}</strong>
                                </SiteAdminAlert>
                            )}
                            <Text>
                                Organization settings apply to all members. User settings override organization
                                settings.
                            </Text>
                        </>
                    }
                />
                {props.isSourcegraphDotCom && props.org.viewerIsMember && props.showOrgDeletion && (
                    <DeleteOrg {...props} />
                )}
            </div>
        ),
    },
    {
        path: '/profile',
        exact: true,
        render: lazyComponent(() => import('./profile/OrgSettingsProfilePage'), 'OrgSettingsProfilePage'),
    },
    {
        path: '/members',
        exact: true,
        render: lazyComponent(() => import('./members/OrgSettingsMembersPage'), 'OrgSettingsMembersPage'),
    },
]
