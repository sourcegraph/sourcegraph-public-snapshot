import { mount } from 'enzyme'
import React from 'react'

import { SearchPatternType } from '../../../../graphql-operations'

import { getResultTypeLinks } from './ResultTypeLink'
import { SearchSidebarProps } from './SearchSidebar'

const defaultProps: SearchSidebarProps = {
    caseSensitive: false,
    patternType: SearchPatternType.literal,
    versionContext: undefined,
    selectedSearchContextSpec: 'global',
    query: 'test',
}

describe('SidebarResultTypeSection', () => {
    it('should have correct links when type not apresent', () => {
        const links = getResultTypeLinks(defaultProps)
        expect(
            mount(
                <>
                    {links.map(link => (
                        <span key={link.key}>{link.node}</span>
                    ))}
                </>
            )
        ).toMatchSnapshot()
    })

    it('should have correct links when type already exists in query', () => {
        const links = getResultTypeLinks({ ...defaultProps, query: 'test type:repo' })
        expect(
            mount(
                <>
                    {links.map(link => (
                        <span key={link.key}>{link.node}</span>
                    ))}
                </>
            )
        ).toMatchSnapshot()
    })
})
