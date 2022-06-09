import React, { useCallback } from 'react'

import classNames from 'classnames'
import CodeBracketsIcon from 'mdi-react/CodeBracketsIcon'
import FormatLetterCaseIcon from 'mdi-react/FormatLetterCaseIcon'
import RegexIcon from 'mdi-react/RegexIcon'

import { isErrorLike, isMacPlatform } from '@sourcegraph/common'
import {
    SearchPatternTypeProps,
    CaseSensitivityProps,
    SearchContextProps,
    SearchPatternTypeMutationProps,
    SubmitSearchProps,
} from '@sourcegraph/search'
import { KEYBOARD_SHORTCUT_COPY_FULL_QUERY } from '@sourcegraph/shared/src/keyboardShortcuts/keyboardShortcuts'
import { SearchPatternType } from '@sourcegraph/shared/src/schema'
import { findFilter, FilterKind } from '@sourcegraph/shared/src/search/query/query'
import { appendContextFilter } from '@sourcegraph/shared/src/search/query/transformer'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'

import { CopyQueryButton } from './CopyQueryButton'
import { QueryInputToggle } from './QueryInputToggle'

import styles from './Toggles.module.scss'

export interface TogglesProps
    extends SearchPatternTypeProps,
        SearchPatternTypeMutationProps,
        CaseSensitivityProps,
        SettingsCascadeProps,
        Pick<SearchContextProps, 'selectedSearchContextSpec'>,
        Partial<Pick<SubmitSearchProps, 'submitSearch'>> {
    navbarSearchQuery: string
    className?: string
    showCopyQueryButton?: boolean
    /**
     * If set to false makes all buttons non-actionable. The main use case for
     * this prop is showing the toggles in examples. This is different from
     * being disabled, because the buttons still render normally.
     */
    interactive?: boolean
    /** Comes from JSContext only set in the web app. */
    structuralSearchDisabled?: boolean
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
export const Toggles: React.FunctionComponent<React.PropsWithChildren<TogglesProps>> = (props: TogglesProps) => {
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
        showCopyQueryButton = true,
        structuralSearchDisabled,
    } = props

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
        <div className={classNames(className, styles.toggleContainer)}>
            <QueryInputToggle
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
            {showCopyQueryButton && (
                <>
                    <div className={styles.separator} />
                    <CopyQueryButton
                        fullQuery={fullQuery}
                        keyboardShortcutForFullCopy={KEYBOARD_SHORTCUT_COPY_FULL_QUERY}
                        isMacPlatform={isMacPlatform()}
                        className={classNames(styles.toggle, styles.copyQueryButton)}
                    />
                </>
            )}
        </div>
    )
}
