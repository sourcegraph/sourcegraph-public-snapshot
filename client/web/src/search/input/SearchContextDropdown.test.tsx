import { mount } from 'enzyme'
import React from 'react'
import { ButtonDropdown, DropdownToggle } from 'reactstrap'
import { SearchContextDropdown } from './SearchContextDropdown'

describe('SearchContextDropdown', () => {
    it('should start closed', () => {
        const element = mount(<SearchContextDropdown query="" />)
        const button = element.find(ButtonDropdown)
        expect(button.prop('isOpen')).toBe(false)
    })

    it('should open when toggle event happens', () => {
        const element = mount(<SearchContextDropdown query="" />)
        let button = element.find(ButtonDropdown)
        button.invoke('toggle')?.(new MouseEvent('click') as any)

        button = element.find(ButtonDropdown)
        expect(button.prop('isOpen')).toBe(true)
    })

    it('should close if toggle event happens again', () => {
        const element = mount(<SearchContextDropdown query="" />)
        let button = element.find(ButtonDropdown)
        button.invoke('toggle')?.(new MouseEvent('click') as any)

        button = element.find(ButtonDropdown)
        button.invoke('toggle')?.(new MouseEvent('click') as any)

        button = element.find(ButtonDropdown)
        expect(button.prop('isOpen')).toBe(false)
    })

    it('should be enabled if query is empty', () => {
        const element = mount(<SearchContextDropdown query="" />)
        const dropdown = element.find(DropdownToggle)
        expect(dropdown.prop('disabled')).toBe(false)
        expect(dropdown.prop('data-tooltip')).toBe('')
    })

    it('should be enabled if query does not contain context filter', () => {
        const element = mount(<SearchContextDropdown query="test (repo:foo or repogroup:python)" />)
        const dropdown = element.find(DropdownToggle)
        expect(dropdown.prop('disabled')).toBe(false)
        expect(dropdown.prop('data-tooltip')).toBe('')
    })

    it('should be disabled if query contains context filter', () => {
        const element = mount(<SearchContextDropdown query="test (context:foo or repogroup:python)" />)
        const dropdown = element.find(DropdownToggle)
        expect(dropdown.prop('disabled')).toBe(true)
        expect(dropdown.prop('data-tooltip')).toBe('Overridden by query')
    })
})
