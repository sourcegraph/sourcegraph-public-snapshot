import type { Decorator, Meta, StoryFn } from '@storybook/react'

import { Link } from '@sourcegraph/wildcard'

import { Breadcrumbs } from './Breadcrumbs'
import { WebStory } from './WebStory'

const decorator: Decorator = story => <div className="container mt-3">{story()}</div>

const config: Meta = {
    title: 'web/Breadcrumbs',
    decorators: [decorator],
}

export default config

export const Example: StoryFn = () => (
    <WebStory>
        {webProps => (
            <Breadcrumbs
                {...webProps}
                breadcrumbs={[
                    {
                        depth: 0,
                        breadcrumb: { key: 'home', element: <Link to="/">Home</Link>, divider: null },
                    },
                    {
                        depth: 1,
                        breadcrumb: { key: 'repo_area', element: <Link to="/">Repositories</Link> },
                    },
                    {
                        depth: 2,
                        breadcrumb: {
                            key: 'repo',
                            element: (
                                <Link to="/">
                                    sourcegraph/<span className="font-weight-medium">sourcegraph</span>
                                </Link>
                            ),
                        },
                    },
                    {
                        depth: 3,
                        breadcrumb: {
                            key: 'revision',
                            divider: <span className="mx-1">@</span>,
                            element: <span className="text-muted">fb/my-branch</span>,
                        },
                    },
                    {
                        depth: 4,
                        breadcrumb: { key: 'directory1', element: <Link to="/">path</Link> },
                    },
                    {
                        depth: 5,
                        breadcrumb: {
                            key: 'directory2',
                            divider: <span className="mx-1">/</span>,
                            element: <Link to="/">to</Link>,
                        },
                    },
                    {
                        depth: 6,
                        breadcrumb: {
                            key: 'fileName',
                            divider: <span className="mx-1">/</span>,
                            element: <Link to="/">file.tsx</Link>,
                        },
                    },
                ]}
            />
        )}
    </WebStory>
)

Example.parameters = {
    design: {
        type: 'figma',
        url: 'https://www.figma.com/file/BkY8Ak997QauG0Iu2EqArv/Sourcegraph-Components?node-id=230%3A83',
    },
}
