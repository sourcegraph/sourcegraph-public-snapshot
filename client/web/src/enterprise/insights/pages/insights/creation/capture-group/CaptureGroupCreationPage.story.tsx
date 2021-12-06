import { Meta } from '@storybook/react'
import React from 'react'

import { WebStory } from '../../../../../../components/WebStory'

import { CaptureGroupCreationPage } from './CaptureGroupCreationPage'

export default {
    title: 'web/insights/creation-ui/CaptureGroupCreationPage',
    decorators: [story => <WebStory>{() => <div className="p-3 container web-content">{story()}</div>}</WebStory>],
} as Meta

export { CaptureGroupCreationPage }
