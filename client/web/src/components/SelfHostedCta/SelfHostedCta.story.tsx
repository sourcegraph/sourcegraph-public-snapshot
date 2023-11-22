import React from 'react'

import type { Decorator, Meta } from '@storybook/react'

import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Text } from '@sourcegraph/wildcard'

import { WebStory } from '../WebStory'

import { SelfHostedCta, type SelfHostedCtaProps } from './SelfHostedCta'

const decorator: Decorator = story => <WebStory>{() => <div className="container mt-3">{story()}</div>}</WebStory>

const config: Meta = {
    title: 'web/marketing/SelfHostedCta',
    decorators: [decorator],
}

export default config

export const Basic: React.FunctionComponent<React.PropsWithChildren<Partial<SelfHostedCtaProps>>> = (
    props
): JSX.Element => (
    <SelfHostedCta telemetryService={NOOP_TELEMETRY_SERVICE} page="storybook" {...props}>
        <Text className="mb-2">
            <strong>Run Sourcegraph self-hosted for more enterprise features</strong>
        </Text>
        <Text className="mb-2">
            For team oriented functionality, additional code hosts and enterprise only features, install Sourcegraph
            self-hosted.
        </Text>
    </SelfHostedCta>
)

export const CustomDisplay = (): JSX.Element => (
    <Basic contentClassName="font-italic" className="p-2 container border rounded" />
)
