/* eslint-disable no-void */

import React from 'react'

import {
    mdiBottleTonicPlus,
    mdiCardBulletedOutline,
    mdiGit,
    mdiGraphOutline,
    mdiSourceCommit,
    mdiTestTube,
} from '@mdi/js'

import { Recipe } from './components/Recipe'
import { RecipeAction } from './components/RecipeAction'
import { Recipes } from './components/Recipes'

import styles from './CodyActionBarWidget.module.scss'

export interface CodyActionBarWidgetProps {
    repoName: string
}

export const CodyActionBarWidget: React.FC<CodyActionBarWidgetProps> = ({ repoName }) => {
    // TODO: Implment message in progress from SideBarStore.
    const isMessageInProgress = false

    return (
        <Recipes className={styles.actionBarRecipesWrapper}>
            <Recipe title="Explain" icon={mdiCardBulletedOutline}>
                <RecipeAction
                    title="Detailed"
                    onClick={() => {
                        // Implement recipe action.
                    }}
                    disabled={isMessageInProgress}
                />
                <RecipeAction
                    title="High level"
                    onClick={() => {
                        // Implement recipe action.
                    }}
                    disabled={isMessageInProgress}
                />
                <RecipeAction
                    title="Code styles"
                    onClick={() => {
                        // Implement recipe action.
                    }}
                    disabled={isMessageInProgress}
                />
            </Recipe>

            <Recipe
                title="Describe structure"
                icon={mdiGit}
                onClick={() => {
                    // Implement recipe action.
                }}
                disabled={isMessageInProgress}
            />
            <Recipe
                title="Summarize recent commits"
                icon={mdiSourceCommit}
                onClick={() => {
                    // Implement recipe action.
                }}
                disabled={isMessageInProgress}
            />
            <Recipe
                title="List dependencies"
                icon={mdiGraphOutline}
                onClick={() => {
                    // Implement recipe action.
                }}
                disabled={isMessageInProgress}
            />
            <Recipe
                title="Evaluate test coverage"
                icon={mdiTestTube}
                onClick={() => {
                    // Implement recipe action.
                }}
            />
            <Recipe
                title="Analyze health"
                icon={mdiBottleTonicPlus}
                onClick={() => {
                    // Implement recipe action.
                }}
                disabled={isMessageInProgress}
            />
        </Recipes>
    )
}
