import { useEffect, useState } from 'react'

import { mdiSourceRepository } from '@mdi/js'
import type { Decorator, Meta } from '@storybook/react'

import { BrandedStory } from '../../stories'
import { Grid } from '../Grid'
import { Icon } from '../Icon'
import { H1 } from '../Typography'

import { ComboboxOptionText } from './Combobox'
import {
    MultiCombobox,
    MultiComboboxInput,
    MultiComboboxPopover,
    MultiComboboxList,
    MultiComboboxOption,
} from './MultiCombobox'

import styles from './MultiComboboxStory.module.scss'

const decorator: Decorator = story => (
    <BrandedStory>{() => <div className="container mt-3">{story()}</div>}</BrandedStory>
)

const config: Meta = {
    title: 'wildcard/MultiCombobox',
    decorators: [decorator],
    parameters: {},
}

export default config

interface Item {
    id: string
    name: string
}

const DEMO_CONTACT_SUGGESTIONS = [
    { id: 'item_001', name: 'Albert Einstein' },
    { id: 'item_002', name: 'Charles Darwin' },
    { id: 'item_003', name: 'James Cook' },
    { id: 'item_004', name: 'Joan of Arc' },
    { id: 'item_005', name: 'Looooooong looooooooooooooooooooong laaaaaaaast naaaaaaameeeeeeeeee' },
    { id: 'item_006', name: 'Ludwig van Beethoven' },
    { id: 'item_007', name: 'Mahatma Gandhi' },
]

export const MultiComboboxDemo = () => (
    <>
        <H1>MultiCombobox UI</H1>
        <Grid columnCount={2}>
            <MultiComboboxWithPopover />
            <MultiComboboxWithPermanentItems />
            <MultiComboboxWithPlainList />
            <MultiComboboxWithAsyncSearch />
            <MultiComboboxWithCustomOptionUI />
        </Grid>
    </>
)

function MultiComboboxWithPopover() {
    const [selectedItems, setSelectedItems] = useState<Item[]>([
        { id: 'item_004', name: 'Joan of Arc' },
        { id: 'item_006', name: 'Ludwig van Beethoven' },
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
            className="mb-4"
        >
            <MultiComboboxInput placeholder="Search assignee" />
            <small className="text-muted pl-2">Focus the field in order to see option list with suggestions</small>

            <MultiComboboxPopover>
                <MultiComboboxList items={suggestions}>
                    {items =>
                        items.map((item, index) => (
                            <MultiComboboxOption key={item.id} value={item.name} index={index} />
                        ))
                    }
                </MultiComboboxList>
            </MultiComboboxPopover>
        </MultiCombobox>
    )
}

interface MaybePermanentItem extends Item {
    permanent?: boolean
}

function MultiComboboxWithPermanentItems() {
    const [selectedItems, setSelectedItems] = useState<MaybePermanentItem[]>([
        { id: 'item_001', name: 'Albert Einstein' },
        { id: 'item_002', name: 'Charles Darwin' },
        { id: 'item_003', name: 'James Cook', permanent: true },
        { id: 'item_007', name: 'Mahatma Gandhi', permanent: true },
    ])

    const suggestions = DEMO_CONTACT_SUGGESTIONS.filter(
        item => !selectedItems.find(selectedItem => selectedItem.id === item.id)
    )

    return (
        <MultiCombobox
            selectedItems={selectedItems}
            getItemKey={item => item.id}
            getItemName={item => item.name}
            getItemIsPermanent={item => !!item.permanent}
            onSelectedItemsChange={setSelectedItems}
            className="mb-4"
        >
            <MultiComboboxInput placeholder="Search assignee" />
            <small className="text-muted pl-2">
                Selected items can be made permanent. These items will always appear at the beginning of the input list.
            </small>

            <MultiComboboxPopover>
                <MultiComboboxList items={suggestions}>
                    {items =>
                        items.map((item, index) => (
                            <MultiComboboxOption key={item.id} value={item.name} index={index} />
                        ))
                    }
                </MultiComboboxList>
            </MultiComboboxPopover>
        </MultiCombobox>
    )
}

function MultiComboboxWithPlainList() {
    const [selectedItems, setSelectedItems] = useState<Item[]>([
        { id: 'item_004', name: 'Joan of Arc' },
        { id: 'item_006', name: 'Ludwig van Beethoven' },
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
        >
            <MultiComboboxInput placeholder="Search assignee" />
            <small className="text-muted pl-2">Suggestion list could be rendered without popover UI</small>

            <MultiComboboxList items={suggestions} className="mt-2">
                {items => items.map(item => <MultiComboboxOption key={item.id} value={item.name} />)}
            </MultiComboboxList>
        </MultiCombobox>
    )
}

interface ExtendedItem {
    id: string
    name: string
    description: string
}

const DEMO_CONTACT_SUGGESTIONS_WITH_DESCRIPTION: ExtendedItem[] = [
    {
        id: 'item_001',
        name: 'Albert Einstein',
        description:
            'Albert Einstein was a German-born theoretical physicist, widely acknowledged to be one of the greatest and most influential physicists of all time',
    },
    {
        id: 'item_002',
        name: 'Charles Darwin',
        description:
            'Charles Robert Darwin FRS FRGS FLS FZS JP was an English naturalist, geologist, and biologist, widely known for his contributions to evolutionary biology',
    },
    {
        id: 'item_003',
        name: 'James Cook',
        description:
            'James Cook FRS was a British explorer, navigator, cartographer, and captain in the British Royal Navy',
    },
    {
        id: 'item_004',
        name: 'Joan of Arc',
        description: `Joan of Arc is a patron saint of France, honored as a defender of the French
                      nation for her role in the siege of Orléans and her insistence on the coronation of Charles
                      VII of France during the Hundred Years' War.`,
    },
    {
        id: 'item_005',
        name: 'Ludwig van Beethoven',
        description: 'German composer and pianist',
    },
    {
        id: 'item_006',
        name: 'Mahatma Gandhi',
        description:
            'Mohandas Karamchand Gandhi, commonly known as Mahatma Gandhi, was an Indian lawyer, anti-colonial nationalist, and political ethicist ',
    },
]

function MultiComboboxWithCustomOptionUI() {
    const [selectedItems, setSelectedItems] = useState<ExtendedItem[]>([
        DEMO_CONTACT_SUGGESTIONS_WITH_DESCRIPTION[0],
        DEMO_CONTACT_SUGGESTIONS_WITH_DESCRIPTION[1],
    ])

    const suggestions = DEMO_CONTACT_SUGGESTIONS_WITH_DESCRIPTION.filter(
        item => !selectedItems.find(selectedItem => selectedItem.id === item.id)
    )

    return (
        <MultiCombobox
            selectedItems={selectedItems}
            getItemKey={item => item.id}
            getItemName={item => item.name}
            onSelectedItemsChange={setSelectedItems}
        >
            <MultiComboboxInput placeholder="Search assignee" />
            <small className="text-muted pl-2">You can have any layout for suggestion elements</small>

            <MultiComboboxList items={suggestions} className="mt-2">
                {items => items.map((item, index) => <CustomOption key={item.id} item={item} index={index} />)}
            </MultiComboboxList>
        </MultiCombobox>
    )
}

interface CustomOptionProps {
    item: ExtendedItem
    index: number
}

function CustomOption(props: CustomOptionProps) {
    const { item, index } = props

    return (
        <MultiComboboxOption value={item.name} index={index} className={styles.customOption}>
            <span>
                <Icon aria-hidden={true} svgPath={mdiSourceRepository} /> <ComboboxOptionText />
            </span>
            <small className={styles.customOptionDescription}>{item.description}</small>
        </MultiComboboxOption>
    )
}

function MultiComboboxWithAsyncSearch() {
    const [search, setSearch] = useState<string>('')
    const [selectedItems, setSelectedItems] = useState<ExtendedItem[]>([
        DEMO_CONTACT_SUGGESTIONS_WITH_DESCRIPTION[0],
        DEMO_CONTACT_SUGGESTIONS_WITH_DESCRIPTION[1],
    ])

    const { suggestions, loading } = useAssigneeSearch(search)

    const suggestionsWithExcludes = suggestions.filter(
        item => !selectedItems.find(selectedItem => selectedItem.id === item.id)
    )

    return (
        <MultiCombobox
            selectedItems={selectedItems}
            getItemKey={item => item.id}
            getItemName={item => item.name}
            onSelectedItemsChange={setSelectedItems}
        >
            <MultiComboboxInput
                value={search}
                placeholder="Search assignee"
                onChange={event => setSearch(event.target.value)}
                status={loading ? 'loading' : 'initial'}
            />
            <small className="text-muted pl-2">You can connect any search engine on consumer level</small>

            <MultiComboboxPopover className={styles.asyncSearchPopover}>
                <MultiComboboxList items={suggestionsWithExcludes}>
                    {items => items.map((item, index) => <CustomOption key={item.id} item={item} index={index} />)}
                </MultiComboboxList>
            </MultiComboboxPopover>
        </MultiCombobox>
    )
}

function useAssigneeSearch(searchTerm: string): { suggestions: ExtendedItem[]; loading: boolean } {
    const [suggestions, setSuggestions] = useState<ExtendedItem[]>(DEMO_CONTACT_SUGGESTIONS_WITH_DESCRIPTION)
    const [loading, setLoading] = useState(false)

    useEffect(() => {
        let isFresh = true
        setLoading(true)

        fetchAssignees(searchTerm)
            .then(repositories => {
                setLoading(false)
                if (isFresh) {
                    setSuggestions(repositories)
                }
            })
            .catch(console.error)

        return () => {
            isFresh = false
        }
    }, [searchTerm])

    return { suggestions, loading }
}

function fetchAssignees(search: string): Promise<ExtendedItem[]> {
    return new Promise<ExtendedItem[]>(resolve =>
        setTimeout(() => resolve(DEMO_CONTACT_SUGGESTIONS_WITH_DESCRIPTION), 2000)
    ).then(result => result.filter(item => item.name.includes(search)))
}
