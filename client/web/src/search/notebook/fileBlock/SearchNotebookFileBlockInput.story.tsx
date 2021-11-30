import { storiesOf } from '@storybook/react'
import { noop } from 'lodash'
import React from 'react'

import { WebStory } from '../../../components/WebStory'

import { SearchNotebookFileBlockInput } from './SearchNotebookFileBlockInput'

const { add } = storiesOf('web/search/notebook/fileBlock/SearchNotebookFileBlockInput', module).addDecorator(story => (
    <div className="p-3 container">{story()}</div>
))

add('default', () => (
    <WebStory>
        {() => (
            <SearchNotebookFileBlockInput
                placeholder="File block input"
                value="client/web/file.tsx"
                onChange={noop}
                onFocus={noop}
                onBlur={noop}
                isMacPlatform={false}
            />
        )}
    </WebStory>
))

add('valid', () => (
    <WebStory>
        {() => (
            <SearchNotebookFileBlockInput
                placeholder="File block input"
                value="client/web/file.tsx"
                onChange={noop}
                onFocus={noop}
                onBlur={noop}
                isValid={true}
                isMacPlatform={false}
            />
        )}
    </WebStory>
))

add('invalid', () => (
    <WebStory>
        {() => (
            <SearchNotebookFileBlockInput
                placeholder="File block input"
                value="client/web/file.tsx"
                onChange={noop}
                onFocus={noop}
                onBlur={noop}
                isValid={false}
                isMacPlatform={false}
            />
        )}
    </WebStory>
))
