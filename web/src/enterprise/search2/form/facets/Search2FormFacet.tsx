import CloseIcon from 'mdi-react/CloseIcon'
import MenuDownIcon from 'mdi-react/MenuDownIcon'
import React, { useCallback, useState } from 'react'

interface Props {
    title: string
    value: string | null
    className?: string
}

export const Search2FormFacet: React.FunctionComponent<Props> = ({ title, value: initialValue, className = '' }) => {
    const [value, setValue] = useState(initialValue)
    const toggleValue = useCallback(() => setValue(previousValue => (previousValue === null ? initialValue : null)), [
        initialValue,
    ])
    const Icon = value === null ? MenuDownIcon : CloseIcon
    return (
        <section
            className={`Search2FormFacet badge badge-pill font-weight-normal px-2 ${className} ${
                value === null ? 'bg-transparent' : 'badge-primary'
            }`}
        >
            {title}
            <Icon onClick={toggleValue} />
        </section>
    )
}
