import React, { useEffect, useRef, useState } from 'react'

import { mdiChevronDown } from '@mdi/js'
import classNames from 'classnames'

import { Icon, Menu, MenuButton, MenuList } from '@sourcegraph/wildcard'

import styles from './Recipes.module.scss'

export interface RecipeProps {
    title?: string
    icon?: string
    children?: React.ReactNode
}

export function Recipe({ children, title, icon }: RecipeProps): JSX.Element {
    const mainButtonRef = useRef<HTMLButtonElement>(null)
    const [menuWidth, setMenuWidth] = useState<number | undefined>(undefined)

    // Make the width ot the drop down menu the same of the main button.
    useEffect(() => {
        if (mainButtonRef.current) {
            const width = mainButtonRef.current.offsetWidth
            setMenuWidth(width)
        }
    }, [])

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
