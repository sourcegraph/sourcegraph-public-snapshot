import React from 'react'

import classNames from 'classnames'

import { AskCodyIcon } from '@sourcegraph/cody-ui/src/icons/AskCodyIcon'

import styles from './Recipes.module.scss'

export interface RecipesProps {
    className?: string
}

export const Recipes: React.FC<React.PropsWithChildren<RecipesProps>> = ({ children, className }) => (
    <div className={classNames(styles.recipesWrapper, className)}>
        <AskCodyIcon />
        {children}
    </div>
)
