import { mount } from 'enzyme'
import React from 'react'

import { SearchPatternType } from '../../../graphql-operations'

import { getFullQuery, Toggles } from './Toggles'

describe('Toggles', () => {
    describe('getFullQuery', () => {
        test('query without search context, case insensitive, literal', () => {
            expect(getFullQuery('foo', '', false, SearchPatternType.literal)).toBe('foo patternType:literal')
        })

        test('query without search context, case sensitive, literal', () => {
            expect(getFullQuery('foo', '', true, SearchPatternType.literal)).toBe('foo patternType:literal case:yes')
        })

        test('query without search context, case sensitive, regexp', () => {
            expect(getFullQuery('foo', '', true, SearchPatternType.regexp)).toBe('foo patternType:regexp case:yes')
        })

        test('query with search context, case sensitive, regexp', () => {
            expect(getFullQuery('foo', '@user1', true, SearchPatternType.regexp)).toBe(
                'context:@user1 foo patternType:regexp case:yes'
            )
        })

        test('query with existing search context, case sensitive, regexp', () => {
            expect(getFullQuery('context:@user2 foo', '@user1', true, SearchPatternType.regexp)).toBe(
                'context:@user2 foo patternType:regexp case:yes'
            )
        })
    })

    describe('Query input toggle state', () => {
        test('case toggle for case subexpressions', () => {
            expect(
                mount(
                    <Toggles
                        navbarSearchQuery="(case:yes foo) or (case:no bar)"
                        patternType={SearchPatternType.literal}
                        setPatternType={() => undefined}
                        caseSensitive={false}
                        setCaseSensitivity={() => undefined}
                        settingsCascade={{ subjects: null, final: {} }}
                        selectedSearchContextSpec="global"
                    />
                ).find('.test-case-sensitivity-toggle')
            ).toMatchSnapshot()
        })

        test('case toggle for patterntype subexpressions', () => {
            expect(
                mount(
                    <Toggles
                        navbarSearchQuery="(foo patterntype:literal) or (bar patterntype:structural)"
                        patternType={SearchPatternType.literal}
                        setPatternType={() => undefined}
                        caseSensitive={false}
                        setCaseSensitivity={() => undefined}
                        settingsCascade={{ subjects: null, final: {} }}
                        selectedSearchContextSpec="global"
                    />
                ).find('.test-case-sensitivity-toggle')
            ).toMatchSnapshot()
        })

        test('regexp toggle for patterntype subexpressions', () => {
            expect(
                mount(
                    <Toggles
                        navbarSearchQuery="(foo patterntype:literal) or (bar patterntype:structural)"
                        patternType={SearchPatternType.literal}
                        setPatternType={() => undefined}
                        caseSensitive={false}
                        setCaseSensitivity={() => undefined}
                        settingsCascade={{ subjects: null, final: {} }}
                        selectedSearchContextSpec="global"
                    />
                ).find('.test-regexp-toggle')
            ).toMatchSnapshot()
        })
    })
})
