import { ReactElement } from 'react'

import { DecoratorFunction } from '@storybook/addons'
import classNames from 'classnames'

import { ChromaticThemeContext } from '../../hooks/useChromaticTheme'

import styles from './ChromaticThemes.module.scss'

export const withChromaticThemes: DecoratorFunction<ReactElement> = (StoryFunc, { parameters }) => {
    if (parameters?.chromatic?.enableDarkMode) {
        return (
            <>
                <ChromaticThemeContext.Provider value={{ theme: 'light' }}>
                    <div className={classNames('theme-light', styles.themeWrapper)}>
                        <StoryFunc />
                    </div>
                </ChromaticThemeContext.Provider>
                <ChromaticThemeContext.Provider value={{ theme: 'dark' }}>
                    <div className={classNames('theme-dark', styles.themeWrapper)}>
                        <StoryFunc />
                    </div>
                </ChromaticThemeContext.Provider>
            </>
        )
    }

    return <StoryFunc />
}
