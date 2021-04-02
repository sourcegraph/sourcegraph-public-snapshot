import React from 'react'
import { DeleteMonitorModal } from './DeleteMonitorModal'
import { storiesOf } from '@storybook/react'
import { EnterpriseWebStory } from '../../components/EnterpriseWebStory'
import sinon from 'sinon'
import { NEVER } from 'rxjs'
import { mockCodeMonitor } from '../testing/util'

const { add } = storiesOf('web/enterprise/code-monitoring/FormTrigerArea', module).addParameters({
    design: {
        type: 'Figma',
        url:
            'https://www.figma.com/file/Krh7HoQi0GFxtO2k399ZQ6/RFC-227-%E2%80%93-Code-monitoring-actions-and-notifications?node-id=3891%3A41568',
    },
})

add('Empty query', () => {}, { design })
