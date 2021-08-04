import * as React from 'react'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { MagnifyingGlassIcon } from '@sourcegraph/web/src/components/MagnifyingGlassIcon'

import { CtaBanner } from '../../components/CtaBanner'

interface Props extends TelemetryProps {
    className?: string
}

export const SignUpCta: React.FunctionComponent<Props> = ({ className, telemetryService }) => (
    <CtaBanner
        className={className}
        icon={<MagnifyingGlassIcon />}
        title="Improve your workflow"
        bodyText="Sign up to add your code, monitor searches for changes, and access additional search features."
        linkText="Sign up"
        href="/sign-up"
        googleAnalytics={true}
        onClick={() => telemetryService.log('HomePageCTAImproveClicked')}
    />
)
