import React from 'react'

import classNames from 'classnames'

import { RadioButton } from '@sourcegraph/wildcard'

import styles from './RadioButtons.module.scss'

/**
 * Descriptor of a radio button element.
 */
interface RadioButtonNode {
    /**
     * Radio button value.
     */
    id: string | string[] | number

    /**
     * Radio button label.
     */
    label: string

    /**
     * `key` property for radio button wrapper element. If not provided, id.toString() is used instead.
     */
    key?: string | number

    /**
     * Radio button tooltip.
     */
    tooltip?: string
}

interface Props {
    /**
     * An additional class name to set on the root element.
     */
    className?: string

    /**
     * List of radio button elements to render.
     */
    nodes: RadioButtonNode[]

    /**
     * Handler for when a radio button is selected.
     */
    onChange?: (event: React.ChangeEvent<HTMLInputElement>) => void

    /**
     * id of the currently selected RadioButtonNode.
     */
    selected?: string

    /**
     * name of the field that these radio buttons select. Use a unique name for
     * each group of radio buttons in a form.
     */
    name: string
}

/**
 * A row of radio buttons.
 */
export const RadioButtons: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    nodes,
    onChange,
    selected,
    className,
    name,
}) => (
    <div className={classNames(styles.radioButtons, className)}>
        {nodes.map(node => (
            <RadioButton
                key={node.key ? node.key : node.id.toString()}
                id={node.id.toString()}
                title={node.tooltip}
                name={name}
                onChange={onChange}
                value={node.id}
                checked={node.id === selected}
                label={<small className={styles.label}>{node.label}</small>}
            />
        ))}
    </div>
)
