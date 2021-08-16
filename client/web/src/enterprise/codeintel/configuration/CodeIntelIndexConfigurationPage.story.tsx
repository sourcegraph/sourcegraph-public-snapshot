import { storiesOf } from '@storybook/react'
import React from 'react'
import { of } from 'rxjs'

import { EnterpriseWebStory } from '../../components/EnterpriseWebStory'

import { CodeIntelIndexConfigurationPage } from './CodeIntelIndexConfigurationPage'

const { add } = storiesOf('web/codeintel/configuration/CodeIntelIndexConfigurationPage', module)
    .addDecorator(story => <div className="p-3 container">{story()}</div>)
    .addParameters({
        chromatic: {
            viewports: [320, 576, 978, 1440],
        },
    })

add('Empty', () => (
    <EnterpriseWebStory>
        {props => (
            <CodeIntelIndexConfigurationPage
                {...props}
                repo={{ id: '42' }}
                getConfiguration={() =>
                    of({
                        __typename: 'Repository',
                        indexConfiguration: {
                            configuration: '',
                            inferredConfiguration: '',
                        },
                    })
                }
            />
        )}
    </EnterpriseWebStory>
))

add('SavedConfiguration', () => (
    <EnterpriseWebStory>
        {props => (
            <CodeIntelIndexConfigurationPage
                {...props}
                repo={{ id: '42' }}
                getConfiguration={() =>
                    of({
                        __typename: 'Repository',
                        indexConfiguration: {
                            configuration: '{"foo": "bar"}',
                            inferredConfiguration: '',
                        },
                    })
                }
            />
        )}
    </EnterpriseWebStory>
))

add('InferredConfiguration', () => (
    <EnterpriseWebStory>
        {props => (
            <CodeIntelIndexConfigurationPage
                {...props}
                repo={{ id: '42' }}
                getConfiguration={() =>
                    of({
                        __typename: 'Repository',
                        indexConfiguration: {
                            configuration: '{"foo": "bar"}',
                            inferredConfiguration: '{"baz": "bonk"}',
                        },
                    })
                }
            />
        )}
    </EnterpriseWebStory>
))
