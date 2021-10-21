import classNames from 'classnames'
import CodeBracketsIcon from 'mdi-react/CodeBracketsIcon'
import FormatLetterCaseIcon from 'mdi-react/FormatLetterCaseIcon'
import RegexIcon from 'mdi-react/RegexIcon'
import React, { useCallback } from 'react'

import { appendContextFilter } from '@sourcegraph/shared/src/search/query/transformer'
import { findFilter, FilterKind } from '@sourcegraph/shared/src/search/query/validate'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { isErrorLike } from '@sourcegraph/shared/src/util/errors'

import { PatternTypeProps, CaseSensitivityProps, SearchContextProps } from '../..'
import { SearchPatternType } from '../../../graphql-operations'
import { KEYBOARD_SHORTCUT_COPY_FULL_QUERY } from '../../../keyboardShortcuts/keyboardShortcuts'
import { isMacPlatform } from '../../../util'
import { SubmitSearchProps } from '../../helpers'

import { CopyQueryButton } from './CopyQueryButton'
import { QueryInputToggle } from './QueryInputToggle'

export interface TogglesProps
    extends PatternTypeProps,
        CaseSensitivityProps,
        SettingsCascadeProps,
        Pick<SearchContextProps, 'selectedSearchContextSpec'>,
        Partial<Pick<SubmitSearchProps, 'submitSearch'>> {
    navbarSearchQuery: string
    className?: string
}

export const getFullQuery = (
    query: string,
    searchContextSpec: string,
    caseSensitive: boolean,
    patternType: SearchPatternType
): string => {
    const finalQuery = [query, `patternType:${patternType}`, caseSensitive ? 'case:yes' : '']
        .filter(queryPart => !!queryPart)
        .join(' ')
    return appendContextFilter(finalQuery, searchContextSpec)
}

/**
 * The toggles displayed in the query input.
 */
export const Toggles: React.FunctionComponent<TogglesProps> = (props: TogglesProps) => {
    const {
        navbarSearchQuery,
        patternType,
        setPatternType,
        caseSensitive,
        setCaseSensitivity,
        settingsCascade,
        className,
        selectedSearchContextSpec,
        submitSearch,
    } = props

    const structuralSearchDisabled = window.context?.experimentalFeatures?.structuralSearch === 'disabled'

    const submitOnToggle = useCallback(
        (args: { newPatternType: SearchPatternType } | { newCaseSensitivity: boolean }): void => {
            submitSearch?.({
                source: 'filter',
                patternType: 'newPatternType' in args ? args.newPatternType : patternType,
                caseSensitive: 'newCaseSensitivity' in args ? args.newCaseSensitivity : caseSensitive,
                activation: undefined,
            })
        },
        [caseSensitive, patternType, submitSearch]
    )

    const toggleCaseSensitivity = useCallback((): void => {
        const newCaseSensitivity = !caseSensitive
        setCaseSensitivity(newCaseSensitivity)
        submitOnToggle({ newCaseSensitivity })
    }, [caseSensitive, setCaseSensitivity, submitOnToggle])

    const toggleRegexp = useCallback((): void => {
        const newPatternType =
            patternType !== SearchPatternType.regexp ? SearchPatternType.regexp : SearchPatternType.literal

        setPatternType(newPatternType)
        submitOnToggle({ newPatternType })
    }, [patternType, setPatternType, submitOnToggle])

    const toggleStructuralSearch = useCallback((): void => {
        const cascadePatternTypeValue =
            settingsCascade.final &&
            !isErrorLike(settingsCascade.final) &&
            (settingsCascade.final['search.defaultPatternType'] as SearchPatternType)

        const defaultPatternType = cascadePatternTypeValue || SearchPatternType.literal

        const newPatternType: SearchPatternType =
            patternType !== SearchPatternType.structural ? SearchPatternType.structural : defaultPatternType

        setPatternType(newPatternType)
        submitOnToggle({ newPatternType })
    }, [patternType, setPatternType, settingsCascade.final, submitOnToggle])

    const fullQuery = getFullQuery(navbarSearchQuery, selectedSearchContextSpec || '', caseSensitive, patternType)

    return (
        <div className={classNames('toggle-container', className)}>
            <QueryInputToggle
                {...props}
                title="Case sensitivity"
                isActive={caseSensitive}
                onToggle={toggleCaseSensitivity}
                icon={FormatLetterCaseIcon}
                className="test-case-sensitivity-toggle"
                activeClassName="test-case-sensitivity-toggle--active"
                disableOn={[
                    {
                        condition: findFilter(navbarSearchQuery, 'case', FilterKind.Subexpression) !== undefined,
                        reason: 'Query already contains one or more case subexpressions',
                    },
                    {
                        condition: findFilter(navbarSearchQuery, 'patterntype', FilterKind.Subexpression) !== undefined,
                        reason:
                            'Query contains one or more patterntype subexpressions, cannot apply global case-sensitivity',
                    },
                    {
                        condition: patternType === SearchPatternType.structural,
                        reason: 'Structural search is always case sensitive',
                    },
                ]}
            />
            <QueryInputToggle
                {...props}
                title="Regular expression"
                isActive={patternType === SearchPatternType.regexp}
                onToggle={toggleRegexp}
                icon={RegexIcon}
                className="test-regexp-toggle"
                activeClassName="test-regexp-toggle--active"
                disableOn={[
                    {
                        condition: findFilter(navbarSearchQuery, 'patterntype', FilterKind.Subexpression) !== undefined,
                        reason: 'Query already contains one or more patterntype subexpressions',
                    },
                ]}
            />
            {!structuralSearchDisabled && (
                <QueryInputToggle
                    {...props}
                    title="Structural search"
                    className="test-structural-search-toggle"
                    activeClassName="test-structural-search-toggle--active"
                    isActive={patternType === SearchPatternType.structural}
                    onToggle={toggleStructuralSearch}
                    icon={CodeBracketsIcon}
                    disableOn={[
                        {
                            condition:
                                findFilter(navbarSearchQuery, 'patterntype', FilterKind.Subexpression) !== undefined,
                            reason: 'Query already contains one or more patterntype subexpressions',
                        },
                    ]}
                />
            )}
            <div className="toggle-container__separator" />
            <CopyQueryButton
                fullQuery={fullQuery}
                keyboardShortcutForFullCopy={KEYBOARD_SHORTCUT_COPY_FULL_QUERY}
                isMacPlatform={isMacPlatform}
                className="toggle-container__toggle toggle-container__copy-query-button"
            />
        </div>
    )
}
