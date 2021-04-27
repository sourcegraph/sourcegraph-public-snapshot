import { mount } from 'enzyme'
import React from 'react'

import { SearchPatternType } from '../../../../graphql-operations'

import { SearchSidebarProps } from './SearchSidebar'
import { getSearchTypeLinks } from './SearchTypeLink'

const defaultProps: Omit<SearchSidebarProps, 'settingsCascade'> = {
    caseSensitive: false,
    patternType: SearchPatternType.literal,
    versionContext: undefined,
    selectedSearchContextSpec: 'global',
    query: 'test',
}

describe('SearchTypeLink', () => {
    it('should have correct links when type not apresent', () => {
        const links = getSearchTypeLinks(defaultProps)
        expect(mount(<>{links}</>)).toMatchSnapshot()
    })

    it('should have correct links when type already exists in query', () => {
        const links = getSearchTypeLinks({ ...defaultProps, query: 'test type:repo' })
        expect(mount(<>{links}</>)).toMatchSnapshot()
    })
})
