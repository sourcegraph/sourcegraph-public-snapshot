import { storiesOf } from '@storybook/react'
import { radios } from '@storybook/addon-knobs'
import React from 'react'
import webStyles from '../../../../enterprise.scss'
import { Tooltip } from '../../../../components/tooltip/Tooltip'
import { HiddenExternalChangesetNode } from './HiddenExternalChangesetNode'
import { addHours } from 'date-fns'
import { ChangesetExternalState } from '../../../../graphql-operations'

const { add } = storiesOf('web/campaigns/HiddenExternalChangesetNode', module).addDecorator(story => {
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

add('All external states', () => {
    const now = new Date()
    return (
        <>
            {Object.values(ChangesetExternalState).map((externalState, index) => (
                <HiddenExternalChangesetNode
                    key={index}
                    node={{
                        id: 'somechangeset',
                        updatedAt: now.toISOString(),
                        nextSyncAt: addHours(now, 1).toISOString(),
                        externalState,
                    }}
                />
            ))}
        </>
    )
})
