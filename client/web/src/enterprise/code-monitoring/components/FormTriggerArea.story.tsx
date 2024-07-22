import type { Meta, StoryFn } from '@storybook/react'
import sinon from 'sinon'

import { H2 } from '@sourcegraph/wildcard'

import { WebStory } from '../../../components/WebStory'

import { FormTriggerArea } from './FormTriggerArea'

import codeMonitorFormStyles from './CodeMonitorForm.module.scss'

const config: Meta = {
    title: 'web/enterprise/code-monitoring/FormTrigerArea',
    parameters: {
        design: {
            type: 'Figma',
            url: 'https://www.figma.com/file/Krh7HoQi0GFxtO2k399ZQ6/RFC-227-%E2%80%93-Code-monitoring-actions-and-notifications?node-id=3891%3A41568',
        },
    },
}

export default config

export const FormTrigerArea: StoryFn = () => (
    <WebStory>
        {props => (
            <>
                <H2>Closed, empty query</H2>
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

                <H2>Open, empty query</H2>
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

                <H2>Open, partially valid query</H2>
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

                <H2>Open, fully valid query</H2>
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
)

FormTrigerArea.storyName = 'FormTrigerArea'
