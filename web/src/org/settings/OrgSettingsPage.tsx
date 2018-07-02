import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { PageTitle } from '../../components/PageTitle'
import { SettingsPage } from '../../settings/SettingsPage'
import { SiteAdminAlert } from '../../site-admin/SiteAdminAlert'
import { OrgAreaPageProps } from '../area/OrgArea'

interface Props extends OrgAreaPageProps, RouteComponentProps<{}> {
    isLightTheme: boolean
}

/** Displays a page for editing organization settings. */
export const OrgSettingsPage: React.SFC<Props> = props => (
    <>
        <PageTitle title="Organization settings" />
        <SettingsPage
            subject={props.org}
            description={
                <>
                    {props.authenticatedUser &&
                        props.org.viewerCanAdminister &&
                        !props.org.viewerIsMember && (
                            <SiteAdminAlert className="sidebar__alert">
                                Viewing settings for <strong>{props.org.name}</strong>
                            </SiteAdminAlert>
                        )}
                    <p>Organization settings apply to all members. User settings override organization settings.</p>
                </>
            }
            location={props.location}
            history={props.history}
            isLightTheme={props.isLightTheme}
        />
    </>
)
