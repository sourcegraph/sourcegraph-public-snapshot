import { storiesOf } from '@storybook/react'
import { radios } from '@storybook/addon-knobs'
import React from 'react'
import webStyles from '../../../../enterprise.scss'
import { Tooltip } from '../../../../components/tooltip/Tooltip'
import { ChangesetCheckStatusCell } from './ChangesetCheckStatusCell'
import { ChangesetCheckState } from '../../../../graphql-operations'
import { capitalize } from 'lodash'

const { add } = storiesOf('web/campaigns/ChangesetCheckStatusCell', module).addDecorator(story => {
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

for (const state of Object.values(ChangesetCheckState)) {
    add(capitalize(state), () => <ChangesetCheckStatusCell checkState={state} />)
}
