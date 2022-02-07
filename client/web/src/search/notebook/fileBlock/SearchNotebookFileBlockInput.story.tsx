import { storiesOf } from '@storybook/react'
import { noop } from 'lodash'
import React from 'react'

import { WebStory } from '../../../components/WebStory'

import { SearchNotebookFileBlockInput } from './SearchNotebookFileBlockInput'

const { add } = storiesOf('web/search/notebook/fileBlock/SearchNotebookFileBlockInput', module).addDecorator(story => (
    <div className="container" style={{ padding: '1rem 1rem 8rem 1rem' }}>
        {story()}
    </div>
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

add('default with suggestions', () => (
    <WebStory>
        {() => (
            <SearchNotebookFileBlockInput
                placeholder="File block input"
                value="client/web/file"
                onChange={noop}
                onFocus={noop}
                onBlur={noop}
                isMacPlatform={false}
                suggestions={['client/web/file1.tsx', 'client/web/file2.tsx', 'client/web/file3.tsx']}
                focusInput={true}
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
