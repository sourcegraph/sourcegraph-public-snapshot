import React from 'react'

import classNames from 'classnames'

import { AskCodyIcon } from '@sourcegraph/cody-ui/dist/icons/AskCodyIcon'

import styles from './Recipes.module.scss'

export const Recipes: React.FC<React.PropsWithChildren<{}>> = ({ children }) => (
    <div className={classNames(styles.recipesWrapper)}>
        <AskCodyIcon />
        {children}
    </div>
)
