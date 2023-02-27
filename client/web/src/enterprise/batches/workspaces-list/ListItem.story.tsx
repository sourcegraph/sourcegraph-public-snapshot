import React from 'react'

import { DecoratorFn, Meta, Story } from '@storybook/react'

import { WebStory } from '../../../components/WebStory'
import { mockPreviewWorkspace } from '../batch-spec/batch-spec.mock'

import { Descriptor } from './Descriptor'
import { CachedIcon, ExcludeIcon } from './Icons'
import { ListItem } from './ListItem'

const decorator: DecoratorFn = story => <div className="list-group w-100">{story()}</div>

const config: Meta = {
    title: 'web/batches/workspaces-list/ListItem',
    decorators: [decorator],
}

export default config

const STATUS_INDICATORS: [key: string, icon: React.FunctionComponent<React.PropsWithChildren<unknown>>][] = [
    ['cached', CachedIcon],
    ['exclude', ExcludeIcon],
]

export const Basic: Story = () => (
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
                                name: 'sourcegraph.github.com/some-really-long-organization-name/an-even-longer-repo-name-for-some-reason-that-just-keeps-going',
                                url: 'lol.fake',
                            },
                        })}
                    />
                </ListItem>
            </>
        )}
    </WebStory>
)

export const NonRootPath: Story = () => (
    <WebStory>
        {props => (
            <>
                <ListItem {...props}>
                    <Descriptor workspace={mockPreviewWorkspace(1, { path: 'path/to/workspace' })} />
                </ListItem>
                <ListItem {...props}>
                    <Descriptor
                        workspace={mockPreviewWorkspace(2, {
                            path: 'a/really/deeply/nested/path/that/is/super/long/and/obnoxious/like/it/just/keeps/going/and-what-the-heck-is-this-folder-name-its-just-so-long/path/to/workspace',
                        })}
                    />
                </ListItem>
            </>
        )}
    </WebStory>
)

NonRootPath.storyName = 'non-root path'

export const WithStatusIndicator: Story = () => (
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
)

WithStatusIndicator.storyName = 'with status indicator'

export const WithClickHandler: Story = () => (
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
)

WithClickHandler.storyName = 'with click handler'
