import React from 'react'
import { CodeMonitoringPage } from './CodeMonitoringPage'
import { storiesOf } from '@storybook/react'
import { WebStory } from '../../components/WebStory'

const { add } = storiesOf('web/enterprise/code-monitoring/CodeMonitoringPage', module)

add('Example', () => <WebStory>{props => <CodeMonitoringPage {...props} />}</WebStory>, {
    design: {
        type: 'figma',
        url:
            'https://www.figma.com/file/Krh7HoQi0GFxtO2k399ZQ6/RFC-227-%E2%80%93-Code-monitoring-actions-and-notifications?node-id=246%3A11',
    },
})
