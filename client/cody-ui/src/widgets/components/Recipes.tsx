import React, { useRef } from 'react'

import classNames from 'classnames'

import { AskCodyIcon } from '../../icons/AskCodyIcon'

import { useRecipesResize } from './useRecipesResize'

import styles from './Recipes.module.css'

interface RecipesProps {
    recipes: Recipe[]
    onSelect?: () => void
}

export function Recipes({ recipes, onSelect }: RecipesProps) {
    const containerRef = useRef<HTMLDivElement>(null)

    // TODO: It is necessary? Check later.
    // Only show Recipe components that fit the available width.
    const visibleRecipes = useRecipesResize({
        recipes,
        containerRef,
    })

    return (
        <div className={classNames(styles.recipesWrapper)} ref={containerRef}>
            <AskCodyIcon />
            {recipes.map((recipe, index) => React.cloneElement(recipe, { key: index }))}
        </div>
    )
}
