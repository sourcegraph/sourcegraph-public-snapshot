import { type FunctionComponent, type PropsWithChildren, useState } from 'react'

import classNames from 'classnames'

import { PopoverRoot } from '@sourcegraph/wildcard'
import { ChromaticThemeContext, type ChromaticTheme } from '@sourcegraph/wildcard/src/stories'

import styles from './ChromaticRoot.module.scss'

interface ChromaticRootProps extends ChromaticTheme {}

export const ChromaticRoot: FunctionComponent<PropsWithChildren<ChromaticRootProps>> = props => {
    const { theme, children } = props

    const [rootReference, setElement] = useState<HTMLDivElement | null>(null)
    const themeClass = theme === 'light' ? 'theme-light' : 'theme-dark'

    return (
        <ChromaticThemeContext.Provider value={{ theme }}>
            {/* Required to render `Popover` inside of the `ChromaticRoot` component. */}
            <PopoverRoot.Provider value={{ renderRoot: rootReference }}>
                <div className={classNames(themeClass, styles.themeWrapper)}>
                    {children}

                    <div ref={setElement} />
                </div>
            </PopoverRoot.Provider>
        </ChromaticThemeContext.Provider>
    )
}
