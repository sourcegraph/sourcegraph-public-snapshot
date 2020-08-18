import { storiesOf } from '@storybook/react'
import { radios } from '@storybook/addon-knobs'
import React from 'react'
import webStyles from '../../../../enterprise.scss'
import { Tooltip } from '../../../../components/tooltip/Tooltip'
import { ChangesetLabel } from './ChangesetLabel'

const { add } = storiesOf('web/campaigns/ChangesetLabel', module).addDecorator(story => {
    const theme = radios('Theme', { Light: 'light', Dark: 'dark' }, 'light')
    document.body.classList.toggle('theme-light', theme === 'light')
    document.body.classList.toggle('theme-dark', theme === 'dark')
    return (
        <>
            <Tooltip />
            <style>{webStyles}</style>
            <div className="p-3 container web-content">{story()}</div>
        </>
    )
})

add('Various labels', () => (
    <>
        <ChangesetLabel
            label={{
                text: 'Feature',
                description: 'A feature, some descriptive text',
                color: '93ba13',
            }}
        />
        <ChangesetLabel
            label={{
                text: 'Bug',
                description: 'A bug, some descriptive text',
                color: 'af1302',
            }}
        />
        <ChangesetLabel
            label={{
                text: 'estimate/1d',
                description: 'An estimation, some descriptive text',
                color: 'bfdadc',
            }}
        />
        <ChangesetLabel
            label={{
                text: 'Debt',
                description: 'Some debt, and some descriptive text',
                color: '795548',
            }}
        />
    </>
))
