import { useMemo } from 'react'

import { mdiCardBulletedOutline, mdiProgressPencil, mdiScrewdriver } from '@mdi/js'

import { Recipe } from './components/Recipe'
import { Recipes } from './components/Recipes'

export const CodyRecipesWidget = (): JSX.Element => {
    const recipes = useMemo(
        () => [
            <Recipe title="Explain" icon={mdiCardBulletedOutline} />,
            <Recipe title="Generate" icon={mdiProgressPencil} />,
            <Recipe title="Fix" icon={mdiScrewdriver} />,
        ],
        []
    )

    const recipesWidget = useMemo(() => <Recipes recipes={recipes} />, [recipes])
    return recipesWidget
}
