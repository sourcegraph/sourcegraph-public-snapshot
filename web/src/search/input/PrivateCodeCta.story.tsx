import { storiesOf } from '@storybook/react'
import { PrivateCodeCta } from './PrivateCodeCta'
import webStyles from '../../SourcegraphWebApp.scss'
import React from 'react'
import { radios } from '@storybook/addon-knobs'

const { add } = storiesOf('web/PrivateCodeCta', module).addDecorator(story => {
    const theme = radios('Theme', { Light: 'light', Dark: 'dark' }, 'light')
    document.body.classList.toggle('theme-light', theme === 'light')
    document.body.classList.toggle('theme-dark', theme === 'dark')

    return (
        <>
            <style>{webStyles}</style>
            <div className="p-4">{story()}</div>
        </>
    )
})

add('Private code cta', () => <PrivateCodeCta />, {
    design: {
        type: 'figma',
        url: 'https://www.figma.com/file/BkY8Ak997QauG0Iu2EqArv/Sourcegraph-Components?node-id=420%3A10',
    },
})
