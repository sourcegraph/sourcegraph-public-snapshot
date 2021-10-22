import { storiesOf } from '@storybook/react'
import React from 'react'
import sinon from 'sinon'

import { WebStory } from '../../../components/WebStory'

import codeMonitorFormStyles from './CodeMonitorForm.module.scss'
import { FormTriggerArea } from './FormTriggerArea'

const { add } = storiesOf('web/enterprise/code-monitoring/FormTrigerArea', module).addParameters({
    design: {
        type: 'Figma',
        url:
            'https://www.figma.com/file/Krh7HoQi0GFxtO2k399ZQ6/RFC-227-%E2%80%93-Code-monitoring-actions-and-notifications?node-id=3891%3A41568',
    },
    chromatic: {
        delay: 600, // Delay screenshot for input validation debouncing
        viewports: [720],
    },
})

add('Open, empty query', () => (
    <WebStory>
        {props => (
            <FormTriggerArea
                {...props}
                query=""
                triggerCompleted={false}
                onQueryChange={sinon.fake()}
                setTriggerCompleted={sinon.fake()}
                startExpanded={true}
                cardBtnClassName={codeMonitorFormStyles.cardButton}
                cardLinkClassName={codeMonitorFormStyles.cardLink}
                cardClassName={codeMonitorFormStyles.card}
            />
        )}
    </WebStory>
))

add('Open, partially valid query', () => (
    <WebStory>
        {props => (
            <FormTriggerArea
                {...props}
                query="test type:commit"
                triggerCompleted={false}
                onQueryChange={sinon.fake()}
                setTriggerCompleted={sinon.fake()}
                startExpanded={true}
                cardBtnClassName={codeMonitorFormStyles.cardButton}
                cardLinkClassName={codeMonitorFormStyles.cardLink}
                cardClassName={codeMonitorFormStyles.card}
            />
        )}
    </WebStory>
))

add('Open, fully valid query', () => (
    <WebStory>
        {props => (
            <FormTriggerArea
                {...props}
                query="test type:commit repo:test"
                triggerCompleted={false}
                onQueryChange={sinon.fake()}
                setTriggerCompleted={sinon.fake()}
                startExpanded={true}
                cardBtnClassName={codeMonitorFormStyles.cardButton}
                cardLinkClassName={codeMonitorFormStyles.cardLink}
                cardClassName={codeMonitorFormStyles.card}
            />
        )}
    </WebStory>
))
