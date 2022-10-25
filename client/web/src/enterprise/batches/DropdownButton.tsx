import React, { useCallback, useEffect, useMemo, useState } from 'react'

import { mdiChevronDown } from '@mdi/js'
import VisuallyHidden from '@reach/visually-hidden'

import {
    ProductStatusBadge,
    Button,
    ButtonGroup,
    Menu,
    MenuButton,
    MenuList,
    Position,
    MenuItem,
    MenuDivider,
    H4,
    Text,
    Icon,
} from '@sourcegraph/wildcard'

import styles from './DropdownButton.module.scss'

export interface Action {
    /* The type of action. Used internally. */
    type: string
    /* The button label for the action. */
    buttonLabel: string
    /* Whether or not the action is disabled. */
    disabled?: boolean
    /* The title in the dropdown menu item. */
    dropdownTitle: string
    /* The description in the dropdown menu item. */
    dropdownDescription: string
    /**
     * Invoked when the action is triggered. Either onDone or onCancel need to
     * be called eventually. Can return a JSX.Element to be rendered adjacent to
     * the button (i.e. a modal).
     */
    onTrigger: (onDone: () => void, onCancel: () => void) => Promise<void | JSX.Element> | void | JSX.Element
    /** If set, displays an experimental badge next to the dropdown title. */
    experimental?: boolean
}

export interface Props {
    actions: Action[]
    defaultAction?: number
    disabled?: boolean
    onLabel?: (label: string | undefined) => void
    placeholder?: string
}

export const DropdownButton: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    actions,
    defaultAction,
    disabled,
    onLabel,
    placeholder = 'Select action',
}) => {
    const [isDisabled, setIsDisabled] = useState(!!disabled)

    const [selected, setSelected] = useState<number | undefined>(undefined)
    const selectedAction = useMemo(() => {
        if (actions.length === 1) {
            return actions[0]
        }

        const id = selected !== undefined ? selected : defaultAction
        if (id !== undefined && id >= 0 && id < actions.length) {
            return actions[id]
        }
        return undefined
    }, [actions, defaultAction, selected])

    const onSelectedTypeSelect = useCallback(
        (type: string) => {
            const index = actions.findIndex(action => action.type === type)
            if (index >= 0) {
                setSelected(actions.findIndex(action => action.type === type))
            } else {
                setSelected(undefined)
            }
        },
        [actions, setSelected]
    )

    const [renderedElement, setRenderedElement] = useState<JSX.Element | undefined>()
    const onTriggerAction = useCallback(async () => {
        if (selectedAction === undefined) {
            return
        }

        // Right now, we don't handle onDone or onCancel separately, but we may
        // want to expose this at a later stage.
        setIsDisabled(true)
        const element = await Promise.resolve(
            selectedAction.onTrigger(
                () => {
                    setIsDisabled(false)
                    setRenderedElement(undefined)
                },
                () => {
                    setIsDisabled(false)
                    setRenderedElement(undefined)
                }
            )
        )
        if (element !== undefined) {
            setRenderedElement(element)
        }
    }, [selectedAction])

    const label = useMemo(() => {
        const label = selectedAction
            ? selectedAction.buttonLabel + (selectedAction.experimental ? ' (Experimental)' : '')
            : undefined

        return label ?? placeholder
    }, [placeholder, selectedAction])

    useEffect(() => {
        if (onLabel) {
            if (selectedAction) {
                onLabel(selectedAction.buttonLabel + (selectedAction.experimental ? ' (Experimental)' : ''))
            }
        }
    })

    return (
        <>
            {renderedElement}
            <Menu>
                <ButtonGroup>
                    <Button
                        className="text-nowrap"
                        onClick={onTriggerAction}
                        disabled={isDisabled || actions.length === 0 || selectedAction === undefined}
                        variant="primary"
                    >
                        {label}
                    </Button>
                    {actions.length > 1 && (
                        <MenuButton variant="primary" className={styles.dropdownButton}>
                            <Icon svgPath={mdiChevronDown} inline={false} aria-hidden={true} />
                            <VisuallyHidden>Actions</VisuallyHidden>
                        </MenuButton>
                    )}
                </ButtonGroup>
                {actions.length > 1 && (
                    <MenuList className={styles.menuList} position={Position.bottomEnd}>
                        {actions.map((action, index) => (
                            <React.Fragment key={action.type}>
                                <DropdownItem action={action} setSelectedType={onSelectedTypeSelect} />
                                {index !== actions.length - 1 && <MenuDivider />}
                            </React.Fragment>
                        ))}
                    </MenuList>
                )}
            </Menu>
        </>
    )
}

interface DropdownItemProps {
    setSelectedType: (type: string) => void
    action: Action
}

const DropdownItem: React.FunctionComponent<React.PropsWithChildren<DropdownItemProps>> = ({
    action,
    setSelectedType,
}) => {
    const onSelect = useCallback(() => {
        setSelectedType(action.type)
    }, [setSelectedType, action.type])
    return (
        <MenuItem className={styles.menuListItem} onSelect={onSelect} disabled={action.disabled}>
            <H4 className="mb-1">
                {action.dropdownTitle}
                {action.experimental && (
                    <>
                        {' '}
                        <ProductStatusBadge status="experimental" as="small" />
                    </>
                )}
            </H4>
            <Text className="text-wrap text-muted mb-0">
                <small>{action.dropdownDescription}</small>
            </Text>
        </MenuItem>
    )
}
