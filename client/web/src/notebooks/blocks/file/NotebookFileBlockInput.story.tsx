import React from 'react'

import { storiesOf } from '@storybook/react'
import { noop } from 'lodash'

import { WebStory } from '../../../components/WebStory'

import { NotebookFileBlockInput } from './NotebookFileBlockInput'

const { add } = storiesOf('web/search/notebooks/blocks/file/NotebookFileBlockInput', module).addDecorator(story => (
    <div className="container" style={{ padding: '1rem 1rem 8rem 1rem' }}>
        {story()}
    </div>
))

add('default', () => (
    <WebStory>
        {() => (
            <NotebookFileBlockInput
                placeholder="File block input"
                value="client/web/file.tsx"
                onChange={noop}
                onFocus={noop}
                onBlur={noop}
            />
        )}
    </WebStory>
))

add('default with suggestions', () => (
    <WebStory>
        {() => (
            <NotebookFileBlockInput
                placeholder="File block input"
                value="client/web/file"
                onChange={noop}
                onFocus={noop}
                onBlur={noop}
                suggestions={['client/web/file1.tsx', 'client/web/file2.tsx', 'client/web/file3.tsx']}
                focusInput={true}
            />
        )}
    </WebStory>
))

add('valid', () => (
    <WebStory>
        {() => (
            <NotebookFileBlockInput
                placeholder="File block input"
                value="client/web/file.tsx"
                onChange={noop}
                onFocus={noop}
                onBlur={noop}
                isValid={true}
            />
        )}
    </WebStory>
))

add('invalid', () => (
    <WebStory>
        {() => (
            <NotebookFileBlockInput
                placeholder="File block input"
                value="client/web/file.tsx"
                onChange={noop}
                onFocus={noop}
                onBlur={noop}
                isValid={false}
            />
        )}
    </WebStory>
))
