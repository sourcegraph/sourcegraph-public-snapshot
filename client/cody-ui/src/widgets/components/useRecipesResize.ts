import { useEffect, useState, RefObject } from 'react'

interface UseRecipesResizeObserverParams {
    recipes: Recipe[]
    containerRef: RefObject<HTMLDivElement>
}

/**
 * A custom hook that observes the container element for size changes
 * and calculates the visible recipes based on the container width.
 *
 * @param params - An object with the following properties:
 *  - `recipes`: an array of `Recipe` objects to be displayed.
 *  - `containerRef`: a `RefObject` pointing to the container element.
 * @returns An array of `Recipe` objects that fit within the container's width.
 */
export function useRecipesResize({ recipes, containerRef }: UseRecipesResizeObserverParams) {
    const [visibleRecipes, setVisibleRecipes] = useState(recipes)

    useEffect(() => {
        const updateVisibleRecipes = () => {
            if (!containerRef.current) return

            const containerWidth = containerRef.current.offsetWidth
            let newVisibleRecipes: Recipe[] = []
            let newWidth = 22

            recipes.forEach(recipe => {
                const recipeElement = document.createElement('div')
                recipeElement.classList.add('recipeWrapper')
                recipeElement.style.visibility = 'hidden'
                recipeElement.innerHTML = recipe.props.title
                document.body.appendChild(recipeElement)

                const recipeWidth = recipeElement.offsetWidth
                newWidth += recipeWidth + 8

                if (newWidth <= containerWidth) {
                    newVisibleRecipes.push(recipe)
                } else {
                    recipeElement.style.display = 'none'
                }

                document.body.removeChild(recipeElement)
            })

            setVisibleRecipes(newVisibleRecipes)
        }

        updateVisibleRecipes()

        if (containerRef.current) {
            const resizeObserver = new ResizeObserver(updateVisibleRecipes)
            resizeObserver.observe(containerRef.current)
            return () => resizeObserver.disconnect()
        }
    }, [recipes, containerRef])

    return visibleRecipes
}
