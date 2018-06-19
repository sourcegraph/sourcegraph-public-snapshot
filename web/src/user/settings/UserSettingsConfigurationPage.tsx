import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { PageTitle } from '../../components/PageTitle'
import { SettingsConfigurationPage } from '../../settings/SettingsConfigurationPage'
import { UserAreaPageProps } from '../area/UserArea'

interface Props extends UserAreaPageProps, RouteComponentProps<{}> {
    isLightTheme: boolean
}

/** Displays a page for editing user settings. */
export const UserSettingsConfigurationPage: React.SFC<Props> = props => (
    <>
        <PageTitle title="User configuration" />
        <SettingsConfigurationPage
            subject={props.user}
            description={<p>User settings override global and organization settings.</p>}
            location={props.location}
            history={props.history}
            isLightTheme={props.isLightTheme}
        />
    </>
)
