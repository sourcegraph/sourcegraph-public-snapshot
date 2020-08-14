import { storiesOf } from '@storybook/react'
import { radios } from '@storybook/addon-knobs'
import React from 'react'
import webStyles from '../../../../enterprise.scss'
import { Tooltip } from '../../../../components/tooltip/Tooltip'
import {
    ChangesetStatusUnpublished,
    ChangesetStatusProcessing,
    ChangesetStatusClosed,
    ChangesetStatusMerged,
    ChangesetStatusOpen,
    ChangesetStatusDeleted,
    ChangesetStatusError,
} from './ChangesetStatusCell'

const { add } = storiesOf('web/campaigns/ChangesetStatusCell', module).addDecorator(story => {
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

add('Unpublished', () => <ChangesetStatusUnpublished />)
add('Closed', () => <ChangesetStatusClosed />)
add('Merged', () => <ChangesetStatusMerged />)
add('Open', () => <ChangesetStatusOpen />)
add('Deleted', () => <ChangesetStatusDeleted />)
add('Error', () => <ChangesetStatusError />)
add('Processing', () => <ChangesetStatusProcessing />)
