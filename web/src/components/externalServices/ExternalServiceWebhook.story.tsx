import { storiesOf } from '@storybook/react'
import { radios } from '@storybook/addon-knobs'
import React from 'react'
import webStyles from '../../SourcegraphWebApp.scss'
import { Tooltip } from '../tooltip/Tooltip'
import { ExternalServiceWebhook } from './ExternalServiceWebhook'
import { ExternalServiceKind } from '../../graphql-operations'

const { add } = storiesOf('web/External services/ExternalServiceWebhook', module).addDecorator(story => {
    const theme = radios('Theme', { Light: 'light', Dark: 'dark' }, 'light')
    document.body.classList.toggle('theme-light', theme === 'light')
    document.body.classList.toggle('theme-dark', theme === 'dark')
    return (
        <>
            <Tooltip />
            <style>{webStyles}</style>
            <div className="p-3 container">{story()}</div>
        </>
    )
})

add('Bitbucket Server', () => (
    <ExternalServiceWebhook
        externalService={{ webhookURL: 'http://test.test/webhook', kind: ExternalServiceKind.BITBUCKETSERVER }}
    />
))
add('GitLab', () => (
    <ExternalServiceWebhook
        externalService={{ webhookURL: 'http://test.test/webhook', kind: ExternalServiceKind.GITLAB }}
    />
))
add('GitHub', () => (
    <ExternalServiceWebhook
        externalService={{ webhookURL: 'http://test.test/webhook', kind: ExternalServiceKind.GITHUB }}
    />
))
