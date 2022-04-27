import { DecoratorFn, Meta } from '@storybook/react'

import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { WebStory } from '../../../components/WebStory'

import { AboutOrganizationPage } from './AboutOrganizationPage'

const decorator: DecoratorFn = story => <WebStory>{() => <div className="container mt-3">{story()}</div>}</WebStory>

const config: Meta = {
    title: 'web/Organizations/AboutOrganizationPage',
    decorators: [decorator],
}

export default config

export const Basic = (): JSX.Element => <AboutOrganizationPage telemetryService={NOOP_TELEMETRY_SERVICE} />
