import { storiesOf } from '@storybook/react'
import { radios } from '@storybook/addon-knobs'
import React from 'react'
import webStyles from '../../SourcegraphWebApp.scss'
import { Tooltip } from '../tooltip/Tooltip'
import { ExternalServiceCard } from './ExternalServiceCard'
import { fetchExternalService as _fetchExternalService } from './backend'
import { allExternalServices } from './externalServices'

const { add } = storiesOf('web/External services/ExternalServiceCard', module).addDecorator(story => {
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

for (const [name, service] of Object.entries(allExternalServices)) {
    add(name, () => (
        <ExternalServiceCard
            icon={service.icon}
            kind={service.kind}
            title={service.title}
            shortDescription={service.shortDescription}
            to="/test"
        />
    ))
}
