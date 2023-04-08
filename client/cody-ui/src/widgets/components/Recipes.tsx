import React, { useRef } from 'react'

import './Recipes.scss'

import { AskCodyIcon } from '../../icons/AskCodyIcon'

import { useRecipesResize } from './useRecipesResize'

interface RecipesProps {
    recipes: Recipe[]
    onSelect?: () => void
}

export function Recipes({ recipes, onSelect }: RecipesProps) {
    const containerRef = useRef<HTMLDivElement>(null)

    // Only show Recipe components that fit the available width.
    const visibleRecipes = useRecipesResize({
        recipes,
        containerRef,
    })

    return (
        <div className="recipesWrapper" ref={containerRef}>
            <AskCodyIcon />
            {visibleRecipes.map((recipe, index) => React.cloneElement(recipe, { key: index }))}
        </div>
    )
}
