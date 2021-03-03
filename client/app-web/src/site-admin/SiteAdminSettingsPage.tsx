import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import * as GQL from '../../../ui-kit-legacy-shared/src/graphql/schema'
import { PlatformContextProps } from '../../../ui-kit-legacy-shared/src/platform/context'
import { SettingsCascadeProps } from '../../../ui-kit-legacy-shared/src/settings/settings'
import { PageTitle } from '../components/PageTitle'
import { SettingsArea } from '../settings/SettingsArea'
import { ThemeProps } from '../../../ui-kit-legacy-shared/src/theme'
import { TelemetryProps } from '../../../ui-kit-legacy-shared/src/telemetry/telemetryService'
import { AuthenticatedUser } from '../auth'

interface Props
    extends RouteComponentProps<{}>,
        PlatformContextProps,
        SettingsCascadeProps,
        ThemeProps,
        TelemetryProps {
    authenticatedUser: AuthenticatedUser
    site: Pick<GQL.ISite, '__typename' | 'id'>
}

export const SiteAdminSettingsPage: React.FunctionComponent<Props> = props => (
    <>
        <PageTitle title="Global settings" />
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
