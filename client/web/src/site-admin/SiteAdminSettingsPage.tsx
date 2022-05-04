import * as React from 'react'

import { RouteComponentProps } from 'react-router'

import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import * as GQL from '@sourcegraph/shared/src/schema'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'

import { AuthenticatedUser } from '../auth'
import { PageTitle } from '../components/PageTitle'
import { SettingsArea } from '../settings/SettingsArea'

interface Props
    extends RouteComponentProps<{}>,
        PlatformContextProps,
        SettingsCascadeProps,
        ThemeProps,
        TelemetryProps {
    authenticatedUser: AuthenticatedUser
    site: Pick<GQL.ISite, '__typename' | 'id'>
}

export const SiteAdminSettingsPage: React.FunctionComponent<React.PropsWithChildren<Props>> = props => (
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
