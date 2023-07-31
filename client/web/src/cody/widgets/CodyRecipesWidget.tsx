/* eslint-disable no-void */

import React, { useEffect } from 'react'

import { mdiCardBulletedOutline, mdiDotsVertical, mdiProgressPencil, mdiShuffleVariant } from '@mdi/js'

import { TranslateToLanguage } from '@sourcegraph/cody-shared/dist/chat/recipes/translate'

import { eventLogger } from '../../tracking/eventLogger'
import { EventName } from '../../util/constants'
import { ChatEditor } from '../components/ChatEditor'
import { CodeMirrorEditor } from '../components/CodeMirrorEditor'
import { CodyChatStore } from '../useCodyChat'

import { Recipe } from './components/Recipe'
import { RecipeAction } from './components/RecipeAction'
import { Recipes } from './components/Recipes'

interface IProps {
    // TODO: change to Editor type
    editor?: CodeMirrorEditor | ChatEditor
    codyChatStore?: CodyChatStore | null
}

export const CodyRecipesWidget: React.FC<IProps> = ({ editor, codyChatStore }) => {
    useEffect(() => {
        eventLogger.log(EventName.CODY_CHAT_EDITOR_WIDGET_VIEWED)
    }, [])

    if (!codyChatStore) {
        return null
    }

    const { executeRecipe, isMessageInProgress, loaded } = codyChatStore

    if (!loaded) {
        return null
    }

    return (
        <div style={{ position: 'absolute', top: '100px', left: '100px' }}>
            <Recipes>
                <Recipe title="Explain" icon={mdiCardBulletedOutline}>
                    <RecipeAction
                        title="Detailed"
                        onClick={() => void executeRecipe('explain-code-detailed', { scope: { editor } })}
                        disabled={isMessageInProgress}
                    />
                    <RecipeAction
                        title="High level"
                        onClick={() => void executeRecipe('explain-code-high-level', { scope: { editor } })}
                        disabled={isMessageInProgress}
                    />
                </Recipe>

                <Recipe title="Generate" icon={mdiProgressPencil}>
                    <RecipeAction
                        title="A unit test"
                        onClick={() => void executeRecipe('generate-unit-test', { scope: { editor } })}
                        disabled={isMessageInProgress}
                    />
                    <RecipeAction
                        title="A docstring"
                        onClick={() => void executeRecipe('generate-docstring', { scope: { editor } })}
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
                                    scope: { editor },
                                })
                            }
                        />
                    ))}
                </Recipe>

                <Recipe icon={mdiDotsVertical}>
                    <RecipeAction
                        title="Improve variable names"
                        disabled={isMessageInProgress}
                        onClick={() => void executeRecipe('improve-variable-names', { scope: { editor } })}
                    />
                    <RecipeAction
                        title="Smell code"
                        onClick={() => void executeRecipe('find-code-smells', { scope: { editor } })}
                        disabled={isMessageInProgress}
                    />
                </Recipe>
            </Recipes>
        </div>
    )
}
