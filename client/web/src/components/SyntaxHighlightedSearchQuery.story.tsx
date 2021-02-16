import React from 'react'
import { storiesOf } from '@storybook/react'
import { SyntaxHighlightedSearchQuery } from './SyntaxHighlightedSearchQuery'
import { WebStory } from './WebStory'

const { add } = storiesOf('web/SyntaxHighlightedSearchQuery', module).addParameters({
    chromatic: { viewports: [480] },
})

add('Examples', () => (
    <WebStory>
        {() => (
            <p className="web-content">
                <SyntaxHighlightedSearchQuery query="test AND spec" />
                <br />
                <SyntaxHighlightedSearchQuery query="test or spec repo:sourcegraph" />
                <br />
                <SyntaxHighlightedSearchQuery query="test -lang:ts" />
            </p>
        )}
    </WebStory>
))
