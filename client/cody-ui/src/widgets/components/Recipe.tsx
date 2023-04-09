import React, { useEffect, useRef, useState } from 'react'

import { mdiChevronDown } from '@mdi/js'
import classNames from 'classnames'

import { Icon, Menu, MenuButton, MenuItem, MenuList } from '@sourcegraph/wildcard'

import styles from './Recipes.module.css'

export interface RecipeProps {
    title?: string
    icon?: string
    onClick?: () => void
    children?: React.ReactNode
}

export function Recipe({ children, title, icon, onClick }: RecipeProps): JSX.Element {
    const mainButtonRef = useRef<HTMLButtonElement>(null)
    const [menuWidth, setMenuWidth] = useState<number | undefined>(undefined)

    const handleClick = () => {
        if (children) {
            // Open sub menu
        } else {
            if (onClick) onClick()
        }
    }

    // Make the width ot the drop down menu the same of the main button.
    useEffect(() => {
        if (mainButtonRef.current) {
            const width = mainButtonRef.current.offsetWidth
            setMenuWidth(width)
        }
    }, [mainButtonRef.current])

    return (
        <Menu>
            <MenuButton
                variant="secondary"
                outline={true}
                className={classNames(styles.recipeWrapper)}
                ref={mainButtonRef}
            >
                {icon && <Icon aria-hidden={true} svgPath={icon} />} {title && title}{' '}
                {children && <Icon aria-hidden={true} svgPath={mdiChevronDown} />}
            </MenuButton>

            {children && (
                <MenuList style={{ minWidth: menuWidth }}>
                    {React.Children.map(children, (child, index) =>
                        React.cloneElement(child as JSX.Element, { key: index })
                    )}
                </MenuList>
            )}
        </Menu>
    )
}
