import { storiesOf } from '@storybook/react'
import React from 'react'
import { Breadcrumbs } from './Breadcrumbs'
import webStyles from '../SourcegraphWebApp.scss'

const { add } = storiesOf('web/Breadcrumbs', module).addDecorator(story => (
    <>
        <style>{webStyles}</style>
        <div className="container mt-3 theme-light">{story()}</div>
    </>
))

add(
    'Example',
    () => (
        <Breadcrumbs
            breadcrumbs={[
                { key: 'home', element: <a href="#">Home</a>, divider: null },
                { key: 'home', element: <a href="#">Repositories</a> },
                {
                    key: 'repo',
                    element: (
                        <a href="#">
                            sourcegraph/<span className="font-weight-semibold">sourcegraph</span>
                        </a>
                    ),
                },
                {
                    key: 'revision',
                    divider: <span className="mx-1 font-weight-semibold">@</span>,
                    element: <span className="text-muted">fb/my-branch</span>,
                },
                { key: 'directory1', element: <a href="#">path</a> },
                {
                    key: 'directory2',
                    divider: <span className="mx-1 font-weight-semibold">/</span>,
                    element: <a href="#">to</a>,
                },
                {
                    key: 'fileName',
                    divider: <span className="mx-1 font-weight-semibold">/</span>,
                    element: <a href="#">file.tsx</a>,
                },
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
