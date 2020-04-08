import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import * as GQL from '../../../shared/src/graphql/schema'
import { PlatformContextProps } from '../../../shared/src/platform/context'
import { SettingsCascadeProps } from '../../../shared/src/settings/settings'
import { PageTitle } from '../components/PageTitle'
import { SettingsArea } from '../settings/SettingsArea'

interface Props extends RouteComponentProps<{}>, PlatformContextProps, SettingsCascadeProps {
    authenticatedUser: GQL.IUser
    isLightTheme: boolean
    site: Pick<GQL.ISite, '__typename' | 'id'>
}

export const SiteAdminSettingsPage: React.FunctionComponent<Props> = props => (
    <>
        <PageTitle title="Site settings" />
        <SettingsArea
            {...props}
            subject={props.site}
            authenticatedUser={props.authenticatedUser}
            extraHeader={
                <p>
                    Global settings apply to all organizations and users. Settings for a user or organization override
                    global settings.
                </p>
            }
        />
    </>
)
