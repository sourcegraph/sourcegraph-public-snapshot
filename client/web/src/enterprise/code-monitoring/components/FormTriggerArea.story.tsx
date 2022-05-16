import { storiesOf } from '@storybook/react'
import sinon from 'sinon'

import { Typography } from '@sourcegraph/wildcard'

import { WebStory } from '../../../components/WebStory'

import { FormTriggerArea } from './FormTriggerArea'

import codeMonitorFormStyles from './CodeMonitorForm.module.scss'

const { add } = storiesOf('web/enterprise/code-monitoring/FormTrigerArea', module).addParameters({
    design: {
        type: 'Figma',
        url:
            'https://www.figma.com/file/Krh7HoQi0GFxtO2k399ZQ6/RFC-227-%E2%80%93-Code-monitoring-actions-and-notifications?node-id=3891%3A41568',
    },
    chromatic: {
        delay: 600, // Delay screenshot for input validation debouncing
        viewports: [720],
        disableSnapshot: false,
    },
})

add('FormTrigerArea', () => (
    <WebStory>
        {props => (
            <>
                <Typography.H2>Closed, empty query</Typography.H2>
                <div className="my-2">
                    <FormTriggerArea
                        {...props}
                        query=""
                        triggerCompleted={false}
                        onQueryChange={sinon.fake()}
                        setTriggerCompleted={sinon.fake()}
                        startExpanded={false}
                        cardBtnClassName={codeMonitorFormStyles.cardButton}
                        cardLinkClassName={codeMonitorFormStyles.cardLink}
                        cardClassName={codeMonitorFormStyles.card}
                        isSourcegraphDotCom={false}
                    />
                </div>

                <Typography.H2>Open, empty query</Typography.H2>
                <div className="my-2">
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
                        isSourcegraphDotCom={false}
                    />
                </div>

                <Typography.H2>Open, partially valid query</Typography.H2>
                <div className="my-2">
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
                        isSourcegraphDotCom={false}
                    />
                </div>

                <Typography.H2>Open, fully valid query</Typography.H2>
                <div className="my-2">
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
                        isSourcegraphDotCom={false}
                    />
                </div>
            </>
        )}
    </WebStory>
))
