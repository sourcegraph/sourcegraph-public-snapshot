import DirectionalSignIcon from '@sourcegraph/icons/lib/DirectionalSign'
import * as React from 'react'
import { HeroPage } from './HeroPage'

export const ErrorNotSupportedPage: React.SFC = () => (
    <HeroPage
        icon={DirectionalSignIcon}
        title="Not Found"
        subtitle="This feature is not enabled in the server configuration."
    />
)
