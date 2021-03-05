import React, { useCallback } from 'react'
import * as H from 'history'
import RegexIcon from 'mdi-react/RegexIcon'
import classNames from 'classnames'
import FormatLetterCaseIcon from 'mdi-react/FormatLetterCaseIcon'
import { PatternTypeProps, CaseSensitivityProps, CopyQueryButtonProps, SearchContextProps } from '../..'
import { SettingsCascadeProps } from '../../../../../shared/src/settings/settings'
import { submitSearch } from '../../helpers'
import { QueryInputToggle } from './QueryInputToggle'
import { isErrorLike } from '../../../../../shared/src/util/errors'
import CodeBracketsIcon from 'mdi-react/CodeBracketsIcon'
import { CopyQueryButton } from './CopyQueryButton'
import { SearchPatternType } from '../../../graphql-operations'
import { VersionContextProps } from '../../../../../shared/src/search/util'
import { appendContextFilter } from '../../../../../shared/src/search/query/transformer'
import { findFilter, FilterKind } from '../../../../../shared/src/search/query/validate'
import { KEYBOARD_SHORTCUT_COPY_FULL_QUERY } from '../../../keyboardShortcuts/keyboardShortcuts'
import { isMacPlatform } from '../../../util'

export interface TogglesProps
    extends PatternTypeProps,
        CaseSensitivityProps,
        SettingsCascadeProps,
        CopyQueryButtonProps,
        VersionContextProps,
        Pick<SearchContextProps, 'showSearchContext' | 'selectedSearchContextSpec'> {
    navbarSearchQuery: string
    history: H.History
    location: H.Location
    hasGlobalQueryBehavior?: boolean
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
        history,
        navbarSearchQuery,
        versionContext,
        hasGlobalQueryBehavior,
        patternType,
        setPatternType,
        caseSensitive,
        setCaseSensitivity,
        settingsCascade,
        className,
        copyQueryButton,
        selectedSearchContextSpec,
    } = props

    const structuralSearchDisabled = window.context?.experimentalFeatures?.structuralSearch === 'disabled'

    const submitOnToggle = useCallback(
        (args: { newPatternType: SearchPatternType } | { newCaseSensitivity: boolean }): void => {
            // Only submit search on toggle when the query input has global behavior (i.e. it's on the main search page
            // or global navbar). Non-global inputs don't have the canonical query and need more context, making
            // submit on-toggle undesirable. Also, only submit on toggle only when the query is non-empty.
            const searchQueryNotEmpty = navbarSearchQuery !== ''
            const shouldSubmitSearch = hasGlobalQueryBehavior && searchQueryNotEmpty
            if (shouldSubmitSearch) {
                const activation = undefined
                const source = 'filter'
                const newPatternType = 'newPatternType' in args ? args.newPatternType : patternType
                const newCaseSensitive = 'newCaseSensitivity' in args ? args.newCaseSensitivity : caseSensitive
                submitSearch({
                    history,
                    query: navbarSearchQuery,
                    source,
                    patternType: newPatternType,
                    caseSensitive: newCaseSensitive,
                    versionContext,
                    activation,
                    selectedSearchContextSpec,
                })
            }
        },
        [
            caseSensitive,
            hasGlobalQueryBehavior,
            history,
            navbarSearchQuery,
            patternType,
            versionContext,
            selectedSearchContextSpec,
        ]
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
                className="toggle-container__regexp-button test-regexp-toggle"
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
            {copyQueryButton && (
                <>
                    <div className="toggle-container__separator" />
                    <CopyQueryButton
                        fullQuery={fullQuery}
                        keyboardShortcutForFullCopy={KEYBOARD_SHORTCUT_COPY_FULL_QUERY}
                        isMacPlatform={isMacPlatform}
                        className="toggle-container__toggle toggle-container__copy-query-button"
                    />
                </>
            )}
        </div>
    )
}
