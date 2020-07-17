import { storiesOf } from '@storybook/react'
import { PrivateCodeCta } from './PrivateCodeCta'
import webStyles from '../../SourcegraphWebApp.scss'
import React from 'react'

const { add } = storiesOf('web/PrivateCodeCta', module).addDecorator(story => (
    <>
        <style>{webStyles}</style>
        <div className="p-4">{story()}</div>
    </>
))

add('Private code cta', () => <PrivateCodeCta />)
