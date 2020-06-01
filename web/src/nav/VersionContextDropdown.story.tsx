import * as H from 'history'
import { storiesOf } from '@storybook/react'
import React from 'react'
import { VersionContextDropdown, VersionContextDropdownProps } from './VersionContextDropdown'
import webMainStyles from '../SourcegraphWebApp.scss'
import { subTypeOf } from '../../../shared/src/util/types'
import { SearchPatternType } from '../../../shared/src/graphql/schema'
import { action } from '@storybook/addon-actions'

const { add } = storiesOf('web/VersionContextDropdown', module).addDecorator(story => (
    <>
        <style>{webMainStyles}</style>
        <div className="theme-light">{story()}</div>
    </>
))

const setVersionContext = action('setVersionContext')
const history = H.createMemoryHistory({ keyLength: 0 })
const commonProps = subTypeOf<Partial<VersionContextDropdownProps>>()({
    alwaysShow: true,
    history,
    // Make sure the dropdown is not rendered outside the theme-light container
    portal: false,
    caseSensitive: false,
    patternType: SearchPatternType.literal,
    availableVersionContexts: [
        { name: 'test 1', description: 'test 1', revisions: [{ rev: 'test', repo: 'github.com/test/test' }] },
        { name: 'test 2', description: 'test 2', revisions: [{ rev: 'test', repo: 'github.com/test/test' }] },
        { name: 'test 3', description: 'test 3', revisions: [{ rev: 'test', repo: 'github.com/test/test' }] },
    ],
    navbarSearchQuery: 'test',
    setVersionContext,
})

add('No context selected', () => <VersionContextDropdown {...commonProps} versionContext={undefined} />)
add('Context selected', () => <VersionContextDropdown {...commonProps} versionContext="test 1" />)
add('Selected context appears at the top of the list', () => (
    <VersionContextDropdown {...commonProps} versionContext="test 3" />
))
