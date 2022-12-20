import { Meta, Story } from '@storybook/react'

import { BrandedStory } from '@sourcegraph/branded/src/components/BrandedStory'
import { H1, H2, Code, Text } from '@sourcegraph/wildcard'

import { AlertType } from '../graphql-operations'

import { GlobalAlert } from './GlobalAlert'

import webStyles from '../SourcegraphWebApp.scss'

const config: Meta = {
    title: 'web/GlobalAlert',

    decorators: [
        story => (
            <BrandedStory styles={webStyles}>{() => <div className="container mt-3">{story()}</div>}</BrandedStory>
        ),
    ],

    parameters: {
        component: GlobalAlert,
        chromatic: {
            disableSnapshot: false,
        },
    },
}

export default config

export const GlobalAlerts: Story = () => (
    <div>
        <H1>Global Alert</H1>
        <Text>
            These alerts map to the <Code>AlertType</Code> returned from the backend API
        </Text>
        <H2>Variants</H2>
        {Object.values(AlertType).map(type => (
            <GlobalAlert key={type} alert={{ message: 'Something happened!', isDismissibleWithKey: null, type }} />
        ))}
        <H2>Dismissible</H2>
        <GlobalAlert
            alert={{ message: 'You can dismiss me', isDismissibleWithKey: 'dismiss-key', type: AlertType.INFO }}
        />
    </div>
)
