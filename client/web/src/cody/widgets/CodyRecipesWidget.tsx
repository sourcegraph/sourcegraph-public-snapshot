import { useMemo, useState, useEffect } from 'react'

import { mdiCardBulletedOutline, mdiDotsVertical, mdiProgressPencil, mdiTranslate } from '@mdi/js'

import { useChatStoreState } from '../../stores/codyChat'

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

    const { onSubmit } = useChatStoreState()

    // Super hacky way to post a message.
    // TODO: Integrate with recipes components and format recipes from there.
    const submit = (): void => {
        onSubmit('Explain the following code at a high level:\n```\n' + selectedCode + '\n```\n')
    }

    const recipesWidget = useMemo(
        () => (
            <Recipes>
                <Recipe title="Explain" icon={mdiCardBulletedOutline}>
                    <RecipeAction title="Detailed" />
                    <RecipeAction title="High level" onClick={submit} />
                </Recipe>

                <Recipe title="Generate" icon={mdiProgressPencil}>
                    <RecipeAction title="A unit test" />
                    <RecipeAction title="A docstring" />
                </Recipe>

                {/* TODO: Load languages from the recipes shared code of VSCode extension. */}
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
