import * as H from 'history'
import { storiesOf } from '@storybook/react'
import { radios } from '@storybook/addon-knobs'
import React from 'react'
import webStyles from '../../../enterprise.scss'
import { Tooltip } from '../../../components/tooltip/Tooltip'
import { ChangesetSpecNode } from './ChangesetSpecNode'
import { visibleChangesetSpecStories } from './VisibleChangesetSpecNode.story'
import { hiddenChangesetSpecStories } from './HiddenChangesetSpecNode.story'

let isLightTheme = true
const { add } = storiesOf('web/campaigns/apply/ChangesetSpecNode', module).addDecorator(story => {
    const theme = radios('Theme', { Light: 'light', Dark: 'dark' }, 'light')
    document.body.classList.toggle('theme-light', theme === 'light')
    document.body.classList.toggle('theme-dark', theme === 'dark')
    isLightTheme = theme === 'light'
    return (
        <>
            <Tooltip />
            <style>{webStyles}</style>
            <div className="p-3 container web-content changeset-spec-list__grid">{story()}</div>
        </>
    )
})

add('Overview', () => {
    const history = H.createMemoryHistory()
    const nodes = [...Object.values(visibleChangesetSpecStories), ...Object.values(hiddenChangesetSpecStories)]
    return (
        <>
            {nodes.map((node, index) => (
                <ChangesetSpecNode
                    key={index}
                    node={node}
                    history={history}
                    location={history.location}
                    isLightTheme={isLightTheme}
                />
            ))}
        </>
    )
})
