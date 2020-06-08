import { storiesOf } from '@storybook/react'
import React from 'react'
import css from 'tagged-template-noop'
import { InfoDropdown } from './InfoDropdown'
import webStyles from '../../SourcegraphWebApp.scss'
import bootstrapStyles from 'bootstrap/scss/bootstrap.scss'

const cssVars = css`
    :root {
        --dropdown-color-border: #cad2e2;
    }
`

const { add } = storiesOf('InfoDropdown', module).addDecorator(story => (
    <>
        <style>{webStyles}</style>
        <style>{bootstrapStyles}</style>
        <style>{cssVars}</style>
        <div>{story()}</div>
    </>
))

add('InfoDropdown', () => (
    <InfoDropdown
        title="Type"
        markdown="Search code (file contents), diffs (added/changed/removed lines in commits), commit messages, or symbols."
    />
))
