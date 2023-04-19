/* eslint-disable no-void */

import React from 'react'

import { mdiCardBulletedOutline, mdiDotsVertical, mdiProgressPencil, mdiTranslate } from '@mdi/js'

import { TranslateToLanguage } from '@sourcegraph/cody-shared/src/chat/recipes/translate'

import { useChatStoreState } from '../../stores/codyChat'

import { Recipe } from './components/Recipe'
import { RecipeAction } from './components/RecipeAction'
import { Recipes } from './components/Recipes'

export const CodyRecipesWidget: React.FC<{}> = () => {
    const { executeRecipe } = useChatStoreState()

    return (
        <Recipes>
            <Recipe title="Explain" icon={mdiCardBulletedOutline}>
                <RecipeAction title="Detailed" onClick={() => void executeRecipe('explain-code-detailed')} />
                <RecipeAction title="High level" onClick={() => void executeRecipe('explain-code-high-level')} />
            </Recipe>

            <Recipe title="Generate" icon={mdiProgressPencil}>
                <RecipeAction title="A unit test" onClick={() => void executeRecipe('generate-unit-test')} />
                <RecipeAction title="A docstring" onClick={() => void executeRecipe('generate-docstring')} />
            </Recipe>

            <Recipe title="Translate" icon={mdiTranslate}>
                {TranslateToLanguage.options.map(language => (
                    <RecipeAction
                        key={language}
                        title={language}
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
                    onClick={() => void executeRecipe('improve-variable-names')}
                />
                <RecipeAction title="Smell code" onClick={() => void executeRecipe('find-code-smells')} />
            </Recipe>
        </Recipes>
    )
}
