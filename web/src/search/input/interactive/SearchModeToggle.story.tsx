import { SearchModeToggle } from './SearchModeToggle'
import * as React from 'react'
import webStyles from '../../../SourcegraphWebApp.scss'
// import toggleStyles from './SearchModeToggle.scss'
import { storiesOf } from '@storybook/react'
import { boolean } from '@storybook/addon-knobs'
import css from 'tagged-template-noop'

const cssVars = css`
    :root {
        --dropdown-link-hover-bg: #f2f4f8;
    }
`
const { add } = storiesOf('web/SearchModeToggle', module).addDecorator(story => (
    <>
        <style>{webStyles}</style>
        <style>{cssVars}</style>
        <div className="theme-light">{story()}</div>
    </>
))

add('SearchModeToggle', () => (
    <SearchModeToggle interactiveSearchMode={boolean('Enabled', true)} toggleSearchMode={() => {}} />
))
