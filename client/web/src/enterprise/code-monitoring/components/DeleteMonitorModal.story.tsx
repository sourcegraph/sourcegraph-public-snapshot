import React from 'react'
import { DeleteMonitorModal } from './DeleteMonitorModal'
import { storiesOf } from '@storybook/react'
import { EnterpriseWebStory } from '../../components/EnterpriseWebStory'
import sinon from 'sinon'
import { NEVER } from 'rxjs'
import { mockCodeMonitor } from '../testing/util'

const { add } = storiesOf('web/enterprise/code-monitoring/DeleteMonitorModal', module)

add(
    'Example',
    () => (
        <EnterpriseWebStory>
            {props => (
                <DeleteMonitorModal
                    {...props}
                    isOpen={true}
                    codeMonitor={mockCodeMonitor.node}
                    toggleDeleteModal={sinon.fake()}
                    deleteCodeMonitor={sinon.fake(() => NEVER)}
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
