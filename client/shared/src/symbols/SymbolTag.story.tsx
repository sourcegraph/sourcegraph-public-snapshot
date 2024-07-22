import type { Decorator, Meta, StoryFn } from '@storybook/react'

import { BrandedStory } from '@sourcegraph/wildcard/src/stories'

import { SymbolKind } from '../graphql-operations'

import { SymbolTag } from './SymbolTag'

const decorator: Decorator = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'shared/SymbolTag',
    parameters: {},
    decorators: [decorator],
}

export default config

const symbolKinds = Object.values(SymbolKind)

export const Default: StoryFn = () => (
    <BrandedStory>
        {() => (
            <div>
                {symbolKinds.map(symbolKind => (
                    <div key={symbolKind}>
                        <SymbolTag kind={symbolKind} />
                    </div>
                ))}
            </div>
        )}
    </BrandedStory>
)
