import { mount } from 'enzyme'
import React from 'react'
import { SearchPatternType } from '../../../../graphql-operations'
import { SearchSidebarProps } from './SearchSidebar'
import { SidebarResultTypeSection } from './SidebarResultTypeSection'

const defaultProps: SearchSidebarProps = {
    caseSensitive: false,
    patternType: SearchPatternType.literal,
    versionContext: undefined,
    selectedSearchContextSpec: 'global',
    query: 'test',
}

describe('SidebarResultTypeSection', () => {
    it('should have correct links when type not present', () => {
        const element = mount(<SidebarResultTypeSection {...defaultProps} />)
        expect(element).toMatchSnapshot()
    })

    it('should have correct links when type already exists in query', () => {
        const element = mount(<SidebarResultTypeSection {...defaultProps} query="test type:repo" />)
        expect(element).toMatchSnapshot()
    })
})
