import { Meta } from '@storybook/react'
import React from 'react'

import { WebStory } from '../../../../components/WebStory'

import { CodeInsightsExamples } from './CodeInsightsExamples'

export default {
    title: 'web/insights/getting-started/CodeInsightExamples',
    decorators: [story => <WebStory>{() => story()}</WebStory>],
} as Meta

export const StandardExample = () => <CodeInsightsExamples />
