import * as H from 'history'
import React, { useEffect, memo, useState } from 'react'
import { PlatformContextProps } from '../../../shared/src/platform/context'
import { SettingsCascadeProps } from '../../../shared/src/settings/settings'
import { PageTitle } from '../components/PageTitle'
import { eventLogger } from '../tracking/eventLogger'
import { ExtensionsAreaRouteContext } from './ExtensionsArea'
import { ExtensionsList } from './ExtensionsList'
import { ExtensionBanner } from './ExtensionBanner'

interface Props
    extends Pick<ExtensionsAreaRouteContext, 'authenticatedUser' | 'subject'>,
        PlatformContextProps<'settings' | 'updateSettings' | 'requestGraphQL'>,
        SettingsCascadeProps {
    location: H.Location
    history: H.History
}

/** A page that displays overview information about the available extensions. */
export const ExtensionsOverviewPage = memo<Props>(props => {
    useEffect(() => eventLogger.logViewEvent('ExtensionsOverview'), [])

    const [showBanner] = useState(true)

    return (
        <>
            <div className="container">
                <PageTitle title="Extensions" />

                <div className="py-3">
                    <ExtensionsList {...props} subject={props.subject} settingsCascade={props.settingsCascade} />
                </div>
            </div>
            {/* TODO: Refactor for better state management, conditional rendering */}
            {showBanner && <ExtensionBanner />}
        </>
    )
})
