import { DecoratorFn, Meta, Story } from '@storybook/react'

import { BrandedStory } from '@sourcegraph/branded/src/components/BrandedStory'

import { SymbolKind } from '../graphql-operations'

import { SymbolTag } from './SymbolTag'

const decorator: DecoratorFn = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'shared/SymbolTag',
    parameters: {
        chromatic: { disableSnapshots: false },
    },
    decorators: [decorator],
}

export default config

const symbolKinds = Object.values(SymbolKind)

export const Default: Story = () => (
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
