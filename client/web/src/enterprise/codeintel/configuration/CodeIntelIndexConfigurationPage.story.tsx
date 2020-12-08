import { storiesOf } from '@storybook/react'
import { SuiteFunction } from 'mocha'
import React from 'react'
import { Observable, of } from 'rxjs'
import { RepositoryIndexConfigurationFields } from '../../../graphql-operations'
import { SourcegraphContext } from '../../../jscontext'
import { EnterpriseWebStory } from '../../components/EnterpriseWebStory'
import { CodeIntelIndexConfigurationPage } from './CodeIntelIndexConfigurationPage'

window.context = {} as SourcegraphContext & SuiteFunction

const { add } = storiesOf('web/Codeintel administration/CodeIntelIndexConfiguration', module).addDecorator(story => (
    <>
        <div className="container">{story()}</div>
    </>
))

const commonProps = {
    repo: { id: '42' },
}

const getConfiguration = (): Observable<RepositoryIndexConfigurationFields> =>
    of({
        __typename: 'Repository' as const,
        indexConfiguration: {
            configuration: '{"foo": "bar"}',
        },
    })

add('WithConfiguration', () => (
    <EnterpriseWebStory>
        {props => <CodeIntelIndexConfigurationPage {...props} {...commonProps} getConfiguration={getConfiguration} />}
    </EnterpriseWebStory>
))
