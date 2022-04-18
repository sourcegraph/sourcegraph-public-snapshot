import React from 'react'
import { NEVER } from 'rxjs'
import { storiesOf } from '@storybook/react'


import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { extensionsController } from '@sourcegraph/shared/src/testing/searchTestHelpers'
import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { ActionItemsBar, useWebActionItems } from './ActionItemsBar'
import webStyles from '../../SourcegraphWebApp.scss'
import { AppRouterContainer } from '../../components/AppRouterContainer'

const LOCATION: H.Location = { hash: '', pathname: '/', search: '', state: undefined }

const PLATFORM_CONTEXT: PlatformContextProps = {
    forceUpdateTooltip: () => undefined,
    settings: NEVER,
}

const { add } = storiesOf('web/extensions/ActionItemsBar', module).addDecorator(story => (
    <>
        <style>{webStyles}</style>
        <AppRouterContainer>
            <div className="container mt-3">{story()}</div>
        </AppRouterContainer>
    </>
))

add('default', () => {
    const { useActionItemsBar } = useWebActionItems()

    return (
        <ActionItemsBar
            location={LOCATION}
            useActionItemsBar={useActionItemsBar}
            extensionsController={extensionsController}
            platformContext={PLATFORM_CONTEXT}
            telemetryService={NOOP_TELEMETRY_SERVICE}
        />
    )
})
