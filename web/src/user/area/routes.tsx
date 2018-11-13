import React from 'react'
import { SettingsArea } from '../../settings/SettingsArea'
import { SiteAdminAlert } from '../../site-admin/SiteAdminAlert'
import { UserAccountArea } from '../account/UserAccountArea'
import { UserAreaRoute } from './UserArea'
import { UserOverviewPage } from './UserOverviewPage'

export const userAreaRoutes: ReadonlyArray<UserAreaRoute> = [
    {
        path: '',
        exact: true,
        // tslint:disable-next-line:jsx-no-lambda
        render: props => <UserOverviewPage {...props} />,
    },
    {
        path: '/settings',
        exact: true,
        // tslint:disable-next-line:jsx-no-lambda
        render: props => (
            <SettingsArea
                {...props}
                subject={props.user}
                isLightTheme={props.isLightTheme}
                extraHeader={
                    <>
                        {props.authenticatedUser &&
                            props.user.id !== props.authenticatedUser.id && (
                                <SiteAdminAlert className="sidebar__alert">
                                    Viewing settings for <strong>{props.user.username}</strong>
                                </SiteAdminAlert>
                            )}
                        <p>User settings override global and organization settings.</p>
                    </>
                }
            />
        ),
    },
    {
        path: '/account',
        // tslint:disable-next-line:jsx-no-lambda
        render: props => (
            <UserAccountArea
                {...props}
                routes={props.userAccountAreaRoutes}
                sideBarItems={props.userAccountSideBarItems}
                isLightTheme={props.isLightTheme}
            />
        ),
    },
]
