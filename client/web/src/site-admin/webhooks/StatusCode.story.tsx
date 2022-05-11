import { number } from '@storybook/addon-knobs'
import { storiesOf } from '@storybook/react'

import { WebStory } from '../../components/WebStory'

import { StatusCode } from './StatusCode'

const { add } = storiesOf('web/site-admin/webhooks/StatusCode', module)
    .addDecorator(story => <div className="p-3 container">{story()}</div>)
    .addParameters({
        chromatic: {
            viewports: [576],
        },
    })

add('success', () => <WebStory>{() => <StatusCode code={number('code', 204, { min: 100, max: 399 })} />}</WebStory>)
add('failure', () => <WebStory>{() => <StatusCode code={number('code', 418, { min: 400, max: 599 })} />}</WebStory>)
