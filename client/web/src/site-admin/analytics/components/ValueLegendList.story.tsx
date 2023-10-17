import type { Decorator, Meta, StoryFn } from '@storybook/react'

import { Container, LoadingSpinner } from '@sourcegraph/wildcard'

import { WebStory } from '../../../components/WebStory'

import { ValueLegendItem, ValueLegendList, type ValueLegendListProps } from './ValueLegendList'

const decorator: Decorator = Story => <Story />

const config: Meta = {
    title: 'web/src/site-admin/analytics/components/ValueLegendList',
    decorators: [decorator],
}

export default config

export const SingleValueLegendItem: StoryFn = () => (
    <WebStory>
        {() => (
            <Container>
                <ValueLegendItem value={12345} description="Single item" tooltip="Here is a tooltip" />
            </Container>
        )}
    </WebStory>
)

SingleValueLegendItem.storyName = 'Single value legend item'

export const ValueLegendListStory: StoryFn = () => {
    const items: ValueLegendListProps['items'] = [
        {
            value: 12345,
            description: 'Single item',
            tooltip: 'Here is a tooltip',
            color: 'var(--orange)',
        },
        {
            value: '101',
            description: 'String value',
            tooltip: 'Here is a tooltip',
            color: 'var(--orange)',
        },
        {
            value: '101',
            secondValue: 1000,
            description: 'Right string double value',
            position: 'right',
            tooltip: 'Here is a tooltip',
            color: 'var(--cyan)',
        },
        {
            value: 42,
            description: 'Right side single',
            tooltip: 'Here is a tooltip',
            position: 'right',
            color: 'var(--cyan)',
        },
        {
            value: 12,
            secondValue: 55555,
            description: 'Double item',
            tooltip: 'Here is a tooltip',
            color: 'var(--orange)',
        },
        {
            value: 13,
            secondValue: 37,
            description: 'Right side',
            tooltip: 'Here is a tooltip',
            position: 'right',
            color: 'var(--cyan)',
        },
        {
            value: <LoadingSpinner />,
            secondValue: 37,
            description: 'Latency-2 :(',
            tooltip: 'Still waiting for your "scalable" backend to respond',
            position: 'right',
            color: 'var(--cyan)',
        },
        {
            value: <LoadingSpinner />,
            description: 'Latency :(',
            tooltip: 'Still waiting for your "scalable" backend to respond',
            color: 'var(--orange)',
        },
        {
            value: 1337,
            secondValue: 3705,
            description: 'Double item with tooltip',
            tooltip: 'Here is a tooltip',
            color: 'var(--orange)',
        },
    ]

    return (
        <WebStory>
            {() => (
                <Container>
                    <ValueLegendList items={items} />
                </Container>
            )}
        </WebStory>
    )
}

ValueLegendListStory.storyName = 'Value legend list with items'
