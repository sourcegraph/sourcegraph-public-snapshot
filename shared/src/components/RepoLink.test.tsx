import React from 'react'
import { RepoLink } from './RepoLink'
import { mount } from 'enzyme'

describe('RepoLink', () => {
    test('renders a link when "to" is set', () => {
        expect(mount(<RepoLink repoName="my/repo" to="http://example.com" />).children()).toMatchSnapshot()
    })

    test('renders a fragment when "to" is null', () => {
        expect(mount(<RepoLink repoName="my/repo" to={null} />).children()).toMatchSnapshot()
    })
})
