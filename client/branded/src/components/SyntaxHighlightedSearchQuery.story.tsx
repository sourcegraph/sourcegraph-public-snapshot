import { storiesOf } from '@storybook/react'
import React from 'react'

import { BrandedStory } from './BrandedStory'
import { SyntaxHighlightedSearchQuery } from './SyntaxHighlightedSearchQuery'

const { add } = storiesOf('web/SyntaxHighlightedSearchQuery', module).addParameters({
    chromatic: { viewports: [480] },
})

add('Examples', () => (
    <BrandedStory>
        {() => (
            <p>
                <SyntaxHighlightedSearchQuery query="test AND spec" />
                <br />
                <SyntaxHighlightedSearchQuery query="test or spec repo:sourcegraph" />
                <br />
                <SyntaxHighlightedSearchQuery query="test -lang:ts" />
            </p>
        )}
    </BrandedStory>
))
