import React from 'react'
import { DeleteMonitorModal } from './DeleteMonitorModal'
import { storiesOf } from '@storybook/react'
import { EnterpriseWebStory } from '../../components/EnterpriseWebStory'
import sinon from 'sinon'
import { FormTriggerArea } from './FormTriggerArea'

const { add } = storiesOf('web/enterprise/code-monitoring/FormTrigerArea', module).addParameters({
    design: {
        type: 'Figma',
        url:
            'https://www.figma.com/file/Krh7HoQi0GFxtO2k399ZQ6/RFC-227-%E2%80%93-Code-monitoring-actions-and-notifications?node-id=3891%3A41568',
    },
    chromatic: {
        delay: 600, // Delay screenshot for input validation debouncing
    },
})

add('Open, empty query', () => {
    return (
        <EnterpriseWebStory>
            {props => (
                <FormTriggerArea
                    {...props}
                    query=""
                    triggerCompleted={false}
                    onQueryChange={sinon.fake()}
                    setTriggerCompleted={sinon.fake()}
                    startExpanded={true}
                />
            )}
        </EnterpriseWebStory>
    )
})

add('Open, partially valid query', () => {
    return (
        <EnterpriseWebStory>
            {props => (
                <FormTriggerArea
                    {...props}
                    query="test type:commit"
                    triggerCompleted={false}
                    onQueryChange={sinon.fake()}
                    setTriggerCompleted={sinon.fake()}
                    startExpanded={true}
                />
            )}
        </EnterpriseWebStory>
    )
})

add('Open, fully valid query', () => {
    return (
        <EnterpriseWebStory>
            {props => (
                <FormTriggerArea
                    {...props}
                    query="test type:commit repo:test"
                    triggerCompleted={false}
                    onQueryChange={sinon.fake()}
                    setTriggerCompleted={sinon.fake()}
                    startExpanded={true}
                />
            )}
        </EnterpriseWebStory>
    )
})
