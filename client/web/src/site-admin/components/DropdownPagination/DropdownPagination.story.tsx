import { useState } from 'react'

import { DecoratorFn, Meta, Story } from '@storybook/react'

import { DropdownPagination } from './DropdownPagination'

const decorator: DecoratorFn = story => <div className="container mt-3 w-50">{story()}</div>

const config: Meta = {
    title: 'wildcard/DropdownPagination',
    component: DropdownPagination,

    decorators: [decorator],
}

export default config

export const Default: Story = () => {
    const [limit, setLimit] = useState(25)
    const [offset, setOffset] = useState(0)

    return (
        <DropdownPagination
            limit={limit}
            offset={offset}
            total={100}
            options={[25, 50, 100]}
            onOffsetChange={setOffset}
            onLimitChange={setLimit}
        />
    )
}
