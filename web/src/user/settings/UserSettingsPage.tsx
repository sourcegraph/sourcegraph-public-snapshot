import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { PageTitle } from '../../components/PageTitle'
import { SettingsPage } from '../../settings/SettingsPage'
import { SiteAdminAlert } from '../../site-admin/SiteAdminAlert'
import { UserAreaPageProps } from '../area/UserArea'

interface Props extends UserAreaPageProps, RouteComponentProps<{}> {
    isLightTheme: boolean
}

/** Displays a page for editing user settings. */
export const UserSettingsPage: React.SFC<Props> = props => (
    <>
        <PageTitle title="User settings" />
        <SettingsPage
            subject={props.user}
            description={
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
            location={props.location}
            history={props.history}
            isLightTheme={props.isLightTheme}
        />
    </>
)
