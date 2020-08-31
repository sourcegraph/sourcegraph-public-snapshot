import { radios } from '@storybook/addon-knobs'
import React from 'react'
import { MemoryRouter, MemoryRouterProps, RouteComponentProps, withRouter } from 'react-router'
import { ThemeProps } from '../../../shared/src/theme'
import webStyles from '../SourcegraphWebApp.scss'
import { Tooltip } from './tooltip/Tooltip'

/**
 * Wrapper component for webapp Storybook stories that provides light theme and react-router props.
 * Takes a render function as children that gets called with the props.
 */
export const WebStory: React.FunctionComponent<
    MemoryRouterProps & {
        children: React.FunctionComponent<ThemeProps & RouteComponentProps<any>>
    }
> = ({ children, ...memoryRouterProps }) => {
    const theme = radios('Theme', { Light: 'light', Dark: 'dark' }, 'light')
    document.body.classList.toggle('theme-light', theme === 'light')
    document.body.classList.toggle('theme-dark', theme === 'dark')
    const Children = withRouter(children)
    return (
        <MemoryRouter {...memoryRouterProps}>
            <Tooltip />
            <Children isLightTheme={theme === 'light'} />
            <style title="Webapp CSS">{webStyles}</style>
        </MemoryRouter>
    )
}
