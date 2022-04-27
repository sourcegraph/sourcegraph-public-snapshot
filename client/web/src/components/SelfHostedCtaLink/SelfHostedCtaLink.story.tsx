import { DecoratorFn, Meta } from '@storybook/react'

import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { WebStory } from '../WebStory'

import { SelfHostedCtaLink } from './SelfHostedCtaLink'

const decorator: DecoratorFn = story => <WebStory>{() => <div className="container mt-3">{story()}</div>}</WebStory>

const config: Meta = {
    title: 'web/markering/SelfHostedCtaLink',
    decorators: [decorator],
}

export default config

export const Basic = (): JSX.Element => <SelfHostedCtaLink telemetryService={NOOP_TELEMETRY_SERVICE} page="storybook" />
