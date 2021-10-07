import { DecoratorFn, Meta, Story } from '@storybook/react'
import React, { useCallback, useState } from 'react'

import { EnterpriseWebStory } from '../../../components/EnterpriseWebStory'

import { MultiCombobox } from './MultiCombobox'

const decorator: DecoratorFn = story => (
    <EnterpriseWebStory>{() => <div className="container mt-3">{story()}</div>}</EnterpriseWebStory>
)

const config: Meta = {
    title: 'insights/MultiCombobox',
    decorators: [decorator],
}

export default config

interface Token {
    id: string
    title: string
}

const INITIAL_VALUES: Token[] = [
    { id: '0001', title: 'Personal' },
    { id: '0002', title: 'Org 1' },
    { id: '0003', title: 'Org 2' },
]

const getTokenKey = (token: Token) => token.id
const getTokenTitle = (token: Token) => token.title

export const MultiComboboxStory: Story = () => {
    const [values, setValues] = useState(INITIAL_VALUES)

    const getSuggestions = useCallback(
        () => [
            { id: '0001', title: 'Custom Personal' },
            { id: '0002', title: 'Org 3' },
            { id: '0003', title: 'Org 4' },
            { id: '0004', title: 'Global' },
        ],
        []
    )

    return (
        <div style={{ width: 500 }}>
            <MultiCombobox
                values={values}
                getTokenKey={getTokenKey}
                getTokenTitle={getTokenTitle}
                getSuggestions={getSuggestions}
                onChange={setValues}
                onSearchChange={value => console.log('SearchValue:', value)}
            />
        </div>
    )
}

MultiComboboxStory.storyName = 'MultiCombobox'
