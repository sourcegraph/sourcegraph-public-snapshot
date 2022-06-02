import { FunctionComponent, PropsWithChildren, ReactElement, useState } from 'react'

import { DecoratorFunction } from '@storybook/addons'
import classNames from 'classnames'

import { PopoverRoot } from '@sourcegraph/wildcard'

import { ChromaticThemeContext } from '../../hooks/useChromaticTheme'

import styles from './ChromaticThemes.module.scss'

export const withChromaticThemes: DecoratorFunction<ReactElement> = (StoryFunc, { parameters }) => {
    if (parameters?.chromatic?.enableDarkMode) {
        return (
            <>
                <ChromaticRoot theme="light">
                    <StoryFunc />
                </ChromaticRoot>

                <ChromaticRoot theme="dark">
                    <StoryFunc />
                </ChromaticRoot>
            </>
        )
    }

    return <StoryFunc />
}

interface ChromaticRootProps {
    theme: 'light' | 'dark'
}

const ChromaticRoot: FunctionComponent<PropsWithChildren<ChromaticRootProps>> = props => {
    const { theme, children } = props

    const [rootReference, setElement] = useState<HTMLDivElement | null>(null)
    const themeClass = theme === 'light' ? 'theme-light' : 'theme-dark'

    return (
        <ChromaticThemeContext.Provider value={{ theme }}>
            <PopoverRoot.Provider value={{ renderRoot: rootReference }}>
                <div className={classNames(themeClass, styles.themeWrapper)}>
                    {children}

                    <div ref={setElement} />
                </div>
            </PopoverRoot.Provider>
        </ChromaticThemeContext.Provider>
    )
}
