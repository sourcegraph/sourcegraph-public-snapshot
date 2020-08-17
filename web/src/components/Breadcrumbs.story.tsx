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
            breadcrumbs={{
                home: {
                    breadcrumb: { key: 'home', element: <a href="#">Home</a>, divider: null },
                    children: {
                        repositories: {
                            breadcrumb: { key: 'repositories', element: <a href="#">Repositories</a> },
                            children: {
                                revision: {
                                    breadcrumb: {
                                        key: 'revision',
                                        divider: <span className="mx-1">@</span>,
                                        element: <span className="text-muted">fb/my-branch</span>,
                                    },
                                    children: {
                                        directory1: {
                                            breadcrumb: { key: 'directory1', element: <a href="#">path</a> },
                                            children: {
                                                directory2: {
                                                    breadcrumb: {
                                                        key: 'directory2',
                                                        divider: <span className="mx-1">/</span>,
                                                        element: <a href="#">to</a>,
                                                    },
                                                    children: {
                                                        fileName: {
                                                            breadcrumb: {
                                                                key: 'fileName',
                                                                divider: <span className="mx-1">/</span>,
                                                                element: <a href="#">file.tsx</a>,
                                                            },
                                                            children: {},
                                                        },
                                                    },
                                                },
                                            },
                                        },
                                    },
                                },
                            },
                        },
                    },
                },
            }}
        />
    ),
    {
        design: {
            type: 'figma',
            url: 'https://www.figma.com/file/BkY8Ak997QauG0Iu2EqArv/Sourcegraph-Components?node-id=230%3A83',
        },
    }
)
