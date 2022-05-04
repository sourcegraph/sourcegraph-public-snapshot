import React from 'react'

import { DecoratorFn, Meta } from '@storybook/react'

import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { WebStory } from '../WebStory'

import { SelfHostedCta, SelfHostedCtaProps } from './SelfHostedCta'

const decorator: DecoratorFn = story => <WebStory>{() => <div className="container mt-3">{story()}</div>}</WebStory>

const config: Meta = {
    title: 'web/markering/SelfHostedCta',
    decorators: [decorator],
}

export default config

export const Basic: React.FunctionComponent<React.PropsWithChildren<Partial<SelfHostedCtaProps>>> = (
    props
): JSX.Element => (
    <SelfHostedCta telemetryService={NOOP_TELEMETRY_SERVICE} page="storybook" {...props}>
        <p className="mb-2">
            <strong>Run Sourcegraph self-hosted for more enterprise features</strong>
        </p>
        <p className="mb-2">
            For team oriented functionality, additional code hosts and enterprise only features, install Sourcegraph
            self-hosted.
        </p>
    </SelfHostedCta>
)

export const CustomDisplay = (): JSX.Element => (
    <Basic contentClassName="font-italic" className="p-2 container border rounded" />
)
