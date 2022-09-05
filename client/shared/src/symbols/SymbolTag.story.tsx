import { DecoratorFn, Meta, Story } from '@storybook/react'

// eslint-disable-next-line no-restricted-imports
import { WebStory } from '@sourcegraph/web/src/components/WebStory'

import { SymbolKind } from '../graphql-operations'

import { SymbolTag } from './SymbolTag'

const decorator: DecoratorFn = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'client/shared/src/symbols/SymbolTag',
    parameters: {
        chromatic: { disableSnapshots: false },
    },
    decorators: [decorator],
}

export default config

const symbolKinds = Object.values(SymbolKind)

export const Default: Story = () => (
    <WebStory>
        {() => (
            <div>
                {symbolKinds.map(symbolKind => (
                    <div key={symbolKind}>
                        <SymbolTag kind={symbolKind} />
                    </div>
                ))}
            </div>
        )}
    </WebStory>
)
