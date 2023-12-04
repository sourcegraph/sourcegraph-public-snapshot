import type { ChangeEvent, PropsWithChildren, ReactElement } from 'react'

import { Label } from '@sourcegraph/wildcard'

import styles from './RadioGroupSection.module.scss'

interface RadioGroupSectionProps {
    name: string
    label: string
    value: string
    checked: boolean
    labelId: string
    className?: string
    onChange: (event: ChangeEvent<HTMLInputElement>) => void
}

export function RadioGroupSection(props: PropsWithChildren<RadioGroupSectionProps>): ReactElement {
    const { name, label, value, checked, labelId, children, onChange } = props

    return (
        <div className={styles.radioGroup}>
            {/*
                Standard wildcard input doesn't provide a simple layout for the radio element,
                in order to have custom layout in the repo control we have to use native input
                with custom styles around spacing and layout
            */}
            {/* eslint-disable-next-line react/forbid-elements */}
            <input
                id={labelId}
                name={name}
                type="checkbox"
                value={value}
                checked={checked}
                className={styles.radioGroupInput}
                onChange={onChange}
            />
            <Label htmlFor={labelId} className={styles.radioGroupLabel}>
                {label}
            </Label>

            {checked && <div className={styles.radioGroupContent}>{children}</div>}
        </div>
    )
}
