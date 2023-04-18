import React, { useRef } from 'react'

import classNames from 'classnames'

import { AskCodyIcon } from '@sourcegraph/cody-ui/src/icons/AskCodyIcon'

import styles from './Recipes.module.scss'

export const Recipes: React.FC<React.PropsWithChildren<{}>> = ({ children }) => {
    const containerRef = useRef<HTMLDivElement>(null)

    return (
        <div className={classNames(styles.recipesWrapper)} ref={containerRef}>
            <AskCodyIcon />
            {children}
        </div>
    )
}
