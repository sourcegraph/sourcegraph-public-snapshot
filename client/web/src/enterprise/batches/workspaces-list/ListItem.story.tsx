import React from 'react'

import { storiesOf } from '@storybook/react'

import { WebStory } from '../../../components/WebStory'
import { mockPreviewWorkspace } from '../batch-spec/batch-spec.mock'

import { Descriptor } from './Descriptor'
import { CachedIcon, ExcludeIcon } from './Icons'
import { ListItem } from './ListItem'

const { add } = storiesOf('web/batches/workspaces-list/ListItem', module).addDecorator(story => (
    <div className="list-group w-100">{story()}</div>
))

add('basic', () => (
    <WebStory>
        {props => (
            <>
                <ListItem {...props}>
                    <Descriptor workspace={mockPreviewWorkspace(1)} />
                </ListItem>
                <ListItem {...props}>
                    <Descriptor workspace={mockPreviewWorkspace(2)} />
                </ListItem>
                <ListItem {...props}>
                    <Descriptor
                        workspace={mockPreviewWorkspace(3, {
                            repository: {
                                __typename: 'Repository',
                                id: 'with-long-name',
                                name:
                                    'sourcegraph.github.com/some-really-long-organization-name/an-even-longer-repo-name-for-some-reason-that-just-keeps-going',
                                url: 'lol.fake',
                            },
                        })}
                    />
                </ListItem>
            </>
        )}
    </WebStory>
))

add('non-root path', () => (
    <WebStory>
        {props => (
            <>
                <ListItem {...props}>
                    <Descriptor workspace={mockPreviewWorkspace(1, { path: 'path/to/workspace' })} />
                </ListItem>
                <ListItem {...props}>
                    <Descriptor
                        workspace={mockPreviewWorkspace(2, {
                            path:
                                'a/really/deeply/nested/path/that/is/super/long/and/obnoxious/like/it/just/keeps/going/and-what-the-heck-is-this-folder-name-its-just-so-long/path/to/workspace',
                        })}
                    />
                </ListItem>
            </>
        )}
    </WebStory>
))

const STATUS_INDICATORS: [key: string, icon: React.FunctionComponent<React.PropsWithChildren<unknown>>][] = [
    ['cached', CachedIcon],
    ['exclude', ExcludeIcon],
]

add('with status indicator', () => (
    <WebStory>
        {props => (
            <>
                {STATUS_INDICATORS.map(([key, Component], index) => (
                    <ListItem {...props} key={key}>
                        <Descriptor workspace={mockPreviewWorkspace(index + 1)} statusIndicator={<Component />} />
                    </ListItem>
                ))}
            </>
        )}
    </WebStory>
))

add('with click handler', () => (
    <WebStory>
        {props => (
            <>
                <ListItem {...props} onClick={() => alert('Clicked workspace 1!')}>
                    <Descriptor workspace={mockPreviewWorkspace(1)} />
                </ListItem>
                <ListItem {...props} onClick={() => alert('Clicked workspace 2!')}>
                    <Descriptor workspace={mockPreviewWorkspace(2)} />
                </ListItem>
            </>
        )}
    </WebStory>
))
