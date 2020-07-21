import { storiesOf } from '@storybook/react'
import { PrivateCodeCta } from './PrivateCodeCta'
import webStyles from '../../SourcegraphWebApp.scss'
import React from 'react'

const { add } = storiesOf('web/PrivateCodeCta', module).addDecorator(story => (
    <>
        <style>{webStyles}</style>
        <div className="theme-light p-4">{story()}</div>
    </>
))

add('Private code cta', () => <PrivateCodeCta />, {
    design: {
        type: 'figma',
        url: 'https://www.figma.com/file/BkY8Ak997QauG0Iu2EqArv/Sourcegraph-Components?node-id=420%3A10',
    },
})
