import { storiesOf } from '@storybook/react'
import React from 'react'
import { Breadcrumbs } from './Breadcrumbs'
import webStyles from '../SourcegraphWebApp.scss'

const { add } = storiesOf('web/Breadcrumbs', module).addDecorator(story => (
    <>
        <style>{webStyles}</style>
        <div>{story()}</div>
    </>
))

add(
    'Breadcrumbs',
    () => (
        <Breadcrumbs
            breadcrumbs={[
                <span className="">Level 2</span>,
                <span className="">Level 3</span>,
                <span className="">Page name</span>,
            ]}
        />
    ),
    {
        design: {
            type: 'figma',
            url: 'https://www.figma.com/file/BkY8Ak997QauG0Iu2EqArv/Sourcegraph-Components?node-id=230%3A83',
        },
    }
)
