/* eslint-disable no-void */

import React from 'react'

import { mdiCardBulletedOutline, mdiDotsVertical, mdiProgressPencil, mdiShuffleVariant } from '@mdi/js'

import { TranslateToLanguage } from '@sourcegraph/cody-shared/src/chat/recipes/translate'

import { useCodySidebar } from '../sidebar/Provider'

import { Recipe } from './components/Recipe'
import { RecipeAction } from './components/RecipeAction'
import { Recipes } from './components/Recipes'

export const CodyRecipesWidget: React.FC<{}> = () => {
    // dirty fix becasue it is rendered under a separate React DOM tree.
    const codySidebarStore = (window as any).codySidebarStore as ReturnType<typeof useCodySidebar>
    if (!codySidebarStore) {
        return null
    }

    const { executeRecipe, isMessageInProgress, loaded } = codySidebarStore

    if (!loaded) {
        return null
    }

    return (
        <Recipes>
            <Recipe title="Explain" icon={mdiCardBulletedOutline}>
                <RecipeAction
                    title="Detailed"
                    onClick={() => void executeRecipe('explain-code-detailed')}
                    disabled={isMessageInProgress}
                />
                <RecipeAction
                    title="High level"
                    onClick={() => void executeRecipe('explain-code-high-level')}
                    disabled={isMessageInProgress}
                />
            </Recipe>

            <Recipe title="Generate" icon={mdiProgressPencil}>
                <RecipeAction
                    title="A unit test"
                    onClick={() => void executeRecipe('generate-unit-test')}
                    disabled={isMessageInProgress}
                />
                <RecipeAction
                    title="A docstring"
                    onClick={() => void executeRecipe('generate-docstring')}
                    disabled={isMessageInProgress}
                />
            </Recipe>

            <Recipe title="Transpile" icon={mdiShuffleVariant}>
                {TranslateToLanguage.options.map(language => (
                    <RecipeAction
                        key={language}
                        title={language}
                        disabled={isMessageInProgress}
                        onClick={() =>
                            void executeRecipe('translate-to-language', {
                                prefilledOptions: [[TranslateToLanguage.options, language]],
                            })
                        }
                    />
                ))}
            </Recipe>

            <Recipe icon={mdiDotsVertical}>
                <RecipeAction
                    title="Improve variable names"
                    disabled={isMessageInProgress}
                    onClick={() => void executeRecipe('improve-variable-names')}
                />
                <RecipeAction
                    title="Smell code"
                    onClick={() => void executeRecipe('find-code-smells')}
                    disabled={isMessageInProgress}
                />
            </Recipe>
        </Recipes>
    )
}
