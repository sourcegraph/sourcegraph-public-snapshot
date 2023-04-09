import React, { useRef } from 'react'

import classNames from 'classnames'

import { AskCodyIcon } from '../../icons/AskCodyIcon'

import { useRecipesResize } from './useRecipesResize'

import styles from './Recipes.module.css'

interface RecipesProps {
    onSelect?: () => void
    children?: React.ReactNode
}

export function Recipes({ children, onSelect }: RecipesProps) {
    const containerRef = useRef<HTMLDivElement>(null)

    // TODO: It is necessary? Check later.
    // Only show Recipe components that fit the available width.
    // const visibleRecipes = useRecipesResize({
    //     recipes,
    //     containerRef,
    // })

    return (
        <div className={classNames(styles.recipesWrapper)} ref={containerRef}>
            <AskCodyIcon />
            {React.Children.map(children, (child, index) => React.cloneElement(child as JSX.Element, { key: index }))}
        </div>
    )
}
