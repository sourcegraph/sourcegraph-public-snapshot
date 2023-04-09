import { useMemo, useState, useEffect } from 'react'

import {
    mdiBulletinBoard,
    mdiCardBulletedOutline,
    mdiClipboardTextClockOutline,
    mdiDotsHorizontal,
    mdiDotsVertical,
    mdiProgressPencil,
    mdiScrewdriver,
    mdiTranslate,
} from '@mdi/js'

import { Recipe } from './components/Recipe'
import { RecipeAction } from './components/RecipeAction'
import { Recipes } from './components/Recipes'

interface CodyRecipesWidgetProps {
    selection?: string
}

export const CodyRecipesWidget = ({ selection }: CodyRecipesWidgetProps): JSX.Element => {
    const [selectedCode, setSelectedCode] = useState<string | undefined>(selection)

    useEffect(() => {
        setSelectedCode(selection)
    }, [selection])

    const recipesWidget = useMemo(
        () => (
            <Recipes>
                <Recipe title="Explain" icon={mdiCardBulletedOutline}>
                    <RecipeAction title="Detailed" />
                    <RecipeAction title="High level" />
                </Recipe>

                <Recipe title="Generate" icon={mdiProgressPencil}>
                    <RecipeAction title="A unit test" />
                    <RecipeAction title="A docstring" />
                </Recipe>

                <Recipe title="Translate" icon={mdiTranslate}>
                    <RecipeAction title="Python" />
                    <RecipeAction title="Java" />
                    <RecipeAction title="Javascript" />
                    <RecipeAction title="Go" />
                    <RecipeAction title="Rust" />
                    <RecipeAction title="Erlang" />
                    <RecipeAction title="TypeScript" />
                    <RecipeAction title="Bash" />
                    <RecipeAction title="C#" />
                    <RecipeAction title="C++" />
                    <RecipeAction title="C" />
                    <RecipeAction title="PHP" />
                    <RecipeAction title="Ruby" />
                    <RecipeAction title="Elm" />
                    <RecipeAction title="Kotlin" />
                    <RecipeAction title="Groovy" />
                    <RecipeAction title="BASIC" />
                    <RecipeAction title="R" />
                    <RecipeAction title="Matlab" />
                    <RecipeAction title="Objective-C" />
                    <RecipeAction title="Swift" />
                    <RecipeAction title="Perl" />
                    <RecipeAction title="Julia" />
                    <RecipeAction title="Fortran" />
                    <RecipeAction title="COBOL" />
                    <RecipeAction title="Lisp" />
                    <RecipeAction title="Haskell" />
                </Recipe>
                {/* <Recipe title="Summarize" icon={mdiClipboardTextClockOutline}>
                    <RecipeAction title="Last 5 items" />
                    <RecipeAction title="Last day" />
                    <RecipeAction title="Last week" />
                </Recipe>
                <Recipe title="Improve names" icon={mdiScrewdriver} /> */}
                <Recipe icon={mdiDotsVertical} />
            </Recipes>
        ),
        []
    )

    return recipesWidget
}
