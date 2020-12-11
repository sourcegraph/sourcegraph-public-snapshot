import React from 'react'
import { ManageCodeMonitorPage } from './ManageCodeMonitorPage'
import { storiesOf } from '@storybook/react'
import { AuthenticatedUser } from '../../auth'
import { EnterpriseWebStory } from '../components/EnterpriseWebStory'
import sinon from 'sinon'
import { NEVER, of } from 'rxjs'
import { mockCodeMonitor } from './testing/util'

const { add } = storiesOf('web/enterprise/code-monitoring/ManageCodeMonitorPage', module)

add(
    'Example',
    () => (
        <EnterpriseWebStory>
            {props => (
                <ManageCodeMonitorPage
                    {...props}
                    authenticatedUser={
                        { id: 'foobar', username: 'alice', email: 'alice@alice.com' } as AuthenticatedUser
                    }
                    updateCodeMonitor={sinon.fake()}
                    fetchCodeMonitor={sinon.fake((id: string) => of(mockCodeMonitor))}
                    deleteCodeMonitor={sinon.fake((id: string) => NEVER)}
                />
            )}
        </EnterpriseWebStory>
    ),
    {
        design: {
            type: 'figma',
            url:
                'https://www.figma.com/file/Krh7HoQi0GFxtO2k399ZQ6/RFC-227-%E2%80%93-Code-monitoring-actions-and-notifications?node-id=246%3A11',
        },
    }
)
