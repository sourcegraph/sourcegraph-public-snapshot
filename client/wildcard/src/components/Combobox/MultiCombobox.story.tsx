import { useState } from 'react'

import { DecoratorFn, Meta } from '@storybook/react'

import { BrandedStory } from '../../stories'

import { MultiCombobox, MultiComboboxInput, MultiComboboxList, MultiComboboxOption } from './MultiCombobox'

const decorator: DecoratorFn = story => (
    <BrandedStory>{() => <div className="container mt-3">{story()}</div>}</BrandedStory>
)

const config: Meta = {
    title: 'wildcard/MultiCombobox',
    decorators: [decorator],
}

export default config

interface Item {
    id: string
    name: string
}

const DEMO_CONTACT_SUGGESTIONS = [
    { id: 'item_001', name: 'Joan of Arc' },
    { id: 'item_002', name: 'Ludwig van Beethoven' },
    { id: 'item_003', name: 'James Cook' },
    { id: 'item_004', name: 'Charles Darwin' },
    { id: 'item_005', name: 'Albert Einstein' },
    { id: 'item_006', name: 'Looooooong looooooooooooooooooooong laaaaaaaast naaaaaaameeeeeeeeee' },
    { id: 'item_007', name: 'Mahatma Gandhi' },
]

export const MultiComboboxDemo = () => {
    const [selectedItems, setSelectedItems] = useState<Item[]>([
        { id: 'item_001', name: 'Joan of Arc' },
        { id: 'item_002', name: 'Ludwig van Beethoven' },
    ])

    const suggestions = DEMO_CONTACT_SUGGESTIONS.filter(
        item => !selectedItems.find(selectedItem => selectedItem.id === item.id)
    )

    return (
        <MultiCombobox
            selectedItems={selectedItems}
            getItemKey={item => item.id}
            getItemName={item => item.name}
            onSelectedItemsChange={setSelectedItems}
            // eslint-disable-next-line @typescript-eslint/ban-ts-comment
            // @ts-ignore
            style={{ width: 500 }}
        >
            <MultiComboboxInput placeholder="Search assignee" status="loading" />

            <MultiComboboxList items={suggestions}>
                {items => items.map(item => <MultiComboboxOption key={item.id} value={item.name} />)}
            </MultiComboboxList>
        </MultiCombobox>
    )
}
