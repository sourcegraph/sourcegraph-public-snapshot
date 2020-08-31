import { storiesOf } from '@storybook/react'
import React from 'react'
import { ExternalServiceCard } from './ExternalServiceCard'
import { fetchExternalService as _fetchExternalService } from './backend'
import { allExternalServices } from './externalServices'
import { WebStory } from '../WebStory'

const { add } = storiesOf('web/External services/ExternalServiceCard', module).addDecorator(story => (
    <div className="p-3 container">{story()}</div>
))

for (const [name, service] of Object.entries(allExternalServices)) {
    add(name, () => (
        <WebStory>
            {() => (
                <ExternalServiceCard
                    icon={service.icon}
                    kind={service.kind}
                    title={service.title}
                    shortDescription={service.shortDescription}
                    to="/test"
                />
            )}
        </WebStory>
    ))
}
