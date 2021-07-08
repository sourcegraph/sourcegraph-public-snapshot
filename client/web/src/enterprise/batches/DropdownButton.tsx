import classNames from 'classnames'
import React, { useCallback, useMemo, useState } from 'react'

import styles from './DropdownButton.module.scss'

export interface Action {
    /* The type of action. Used internally. */
    type: string
    /* The button label for the action. */
    buttonLabel: string
    /* The title in the dropdown menu item. */
    dropdownTitle: string
    /* The description in the dropdown menu item. */
    dropdownDescription: string
    /* Conditionally display the action based on the given query arguments. */
    isAvailable: () => boolean
    /**
     * Invoked when the action is triggered. Either onDone or onCancel need to
     * be called eventually. Can return a JSX.Element to be rendered adjacent to
     * the button (i.e. a modal).
     */
    onTrigger: (onDone: () => void, onCancel: () => void) => Promise<void | JSX.Element>
    /** If set, displays an experimental badge next to the dropdown title. */
    experimental?: boolean
}

export interface Props {
    actions: Action[]
    defaultAction?: number
    disabled?: boolean
    initiallyOpen?: boolean
    placeholder?: string
    tooltip?: string
}

export const DropdownButton: React.FunctionComponent<Props> = ({
    actions,
    defaultAction,
    disabled,
    initiallyOpen,
    placeholder,
    tooltip,
}) => {
    placeholder ??= 'Select action'

    actions = useMemo(() => actions.filter(action => action.isAvailable()), [actions])

    const [isDisabled, setIsDisabled] = useState(!!disabled)

    const [isOpen, setIsOpen] = useState(!!initiallyOpen)
    const toggleIsOpen = useCallback(() => setIsOpen(open => !open), [])

    const [selected, setSelected] = useState<Action | undefined>(() => {
        if (defaultAction !== undefined && defaultAction >= 0 && defaultAction < actions.length) {
            return actions[defaultAction]
        }
        if (actions.length === 1) {
            return actions[0]
        }
        return undefined
    })
    const onSelectedTypeSelect = useCallback(
        (type: string) => {
            setSelected(actions.find(action => action.type === type))
            setIsOpen(false)
        },
        [actions, setIsOpen, setSelected]
    )

    const [renderedElement, setRenderedElement] = useState<JSX.Element | undefined>()
    const onTriggerAction = useCallback(async () => {
        if (selected === undefined) {
            return
        }

        // TODO: can we do something useful with the onDone/onCancel split?
        setIsDisabled(true)
        const element = await selected.onTrigger(
            () => {
                setIsDisabled(false)
                setRenderedElement(undefined)
            },
            () => {
                setIsDisabled(false)
                setRenderedElement(undefined)
            }
        )
        if (element !== undefined) {
            setRenderedElement(element)
        }
    }, [selected, setIsDisabled, setRenderedElement])

    return (
        <>
            {renderedElement}
            <div className="btn-group">
                <button
                    type="button"
                    className="btn btn-primary text-nowrap"
                    onClick={onTriggerAction}
                    disabled={isDisabled || actions.length === 0 || selected === undefined}
                    data-tooltip={tooltip}
                >
                    {selected === undefined ? placeholder : selected.buttonLabel}
                    {selected?.experimental && ' (Experimental)'}
                </button>
                {actions.length > 1 && (
                    <>
                        <button
                            type="button"
                            onClick={toggleIsOpen}
                            className="btn btn-primary dropdown-toggle dropdown-toggle-split"
                        />
                        <div
                            className={classNames(
                                'dropdown-menu dropdown-menu-right',
                                isOpen && 'show',
                                styles.dropdownButtonItem
                            )}
                        >
                            {actions.map((action, index) => (
                                <React.Fragment key={action.type}>
                                    <DropdownItem action={action} setSelectedType={onSelectedTypeSelect} />
                                    {index !== actions.length - 1 && <div className="dropdown-divider" />}
                                </React.Fragment>
                            ))}
                        </div>
                    </>
                )}
            </div>
        </>
    )
}

interface DropdownItemProps {
    setSelectedType: (type: string) => void
    action: Action
}

const DropdownItem: React.FunctionComponent<DropdownItemProps> = ({ action, setSelectedType }) => {
    const onClick = useCallback<React.MouseEventHandler>(() => {
        setSelectedType(action.type)
    }, [setSelectedType, action.type])
    return (
        <div className="dropdown-item">
            <button type="button" className="btn text-left" onClick={onClick}>
                <h4 className="mb-1">
                    {action.dropdownTitle}
                    {action.experimental && (
                        <>
                            {' '}
                            <small className="badge badge-info">Experimental</small>
                        </>
                    )}
                </h4>
                <p className="text-wrap text-muted mb-0">
                    <small>{action.dropdownDescription}</small>
                </p>
            </button>
        </div>
    )
}
