import React, { useCallback } from 'react'

import { mdiCodeBrackets, mdiFormatLetterCase, mdiRegex } from '@mdi/js'
import classNames from 'classnames'

import { SearchPatternType } from '@sourcegraph/shared/src/graphql-operations'
import {
    type CaseSensitivityProps,
    type SearchPatternTypeMutationProps,
    type SubmitSearchProps,
    SearchMode,
    type SearchModeProps,
    type SearchPatternTypeProps,
} from '@sourcegraph/shared/src/search'
import { findFilter, FilterKind } from '@sourcegraph/shared/src/search/query/query'

import { QueryInputToggle } from './QueryInputToggle'
import { SmartSearchToggle } from './SmartSearchToggle'
import { SmartSearchToggleExtended, SearchModes } from './SmartSearchToggleExtended'

import styles from './Toggles.module.scss'

export interface TogglesProps
    extends SearchPatternTypeProps,
        SearchPatternTypeMutationProps,
        CaseSensitivityProps,
        SearchModeProps,
        Partial<Pick<SubmitSearchProps, 'submitSearch'>> {
    navbarSearchQuery: string
    className?: string
    showSmartSearchButton?: boolean
    /**
     * If set to true, the search mode picker will let the user select the new
     * pattern type as a new alternative
     */
    showExtendedPicker?: boolean
    /**
     * If set to false makes all buttons non-actionable. The main use case for
     * this prop is showing the toggles in examples. This is different from
     * being disabled, because the buttons still render normally.
     */
    interactive?: boolean
    /** Comes from JSContext only set in the web app. */
    structuralSearchDisabled?: boolean
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
        searchMode,
        setSearchMode,
        className,
        submitSearch,
        showSmartSearchButton = true,
        showExtendedPicker = false,
        structuralSearchDisabled,
    } = props

    const submitOnToggle = useCallback(
        (
            args:
                | { newPatternType: SearchPatternType }
                | { newCaseSensitivity: boolean }
                | { newPowerUser: boolean }
                | { newSearchMode: SearchMode }
        ): void => {
            submitSearch?.({
                source: 'filter',
                patternType: 'newPatternType' in args ? args.newPatternType : patternType,
                caseSensitive: 'newCaseSensitivity' in args ? args.newCaseSensitivity : caseSensitive,
                searchMode: 'newSearchMode' in args ? args.newSearchMode : searchMode,
            })
        },
        [caseSensitive, patternType, searchMode, submitSearch]
    )

    const toggleCaseSensitivity = useCallback((): void => {
        const newCaseSensitivity = !caseSensitive
        setCaseSensitivity(newCaseSensitivity)
        submitOnToggle({ newCaseSensitivity })
    }, [caseSensitive, setCaseSensitivity, submitOnToggle])

    const toggleRegexp = useCallback((): void => {
        const newPatternType =
            patternType !== SearchPatternType.regexp ? SearchPatternType.regexp : SearchPatternType.standard

        setPatternType(newPatternType)
        submitOnToggle({ newPatternType })
    }, [patternType, setPatternType, submitOnToggle])

    const toggleNewStandard = useCallback((): void => {
        const newPatternType =
            patternType !== SearchPatternType.newStandardRC1
                ? SearchPatternType.newStandardRC1
                : SearchPatternType.standard

        setPatternType(newPatternType)

        // We always want precise mode when switching to the experimental pattern type.
        setSearchMode(SearchMode.Precise)

        submitOnToggle({ newPatternType })
    }, [patternType, setPatternType, submitOnToggle, setSearchMode])

    const toggleStructuralSearch = useCallback((): void => {
        const newPatternType: SearchPatternType =
            patternType !== SearchPatternType.structural ? SearchPatternType.structural : SearchPatternType.standard

        setPatternType(newPatternType)
        submitOnToggle({ newPatternType })
    }, [patternType, setPatternType, submitOnToggle])

    const onSelectSmartSearch = useCallback(
        (enabled: boolean): void => {
            const newSearchMode: SearchMode = enabled ? SearchMode.SmartSearch : SearchMode.Precise

            // Disable the experimental pattern type the user activates smart search
            if (patternType === SearchPatternType.newStandardRC1) {
                setPatternType(SearchPatternType.standard)
            }

            setSearchMode(newSearchMode)
            submitOnToggle({ newSearchMode })
        },
        [setSearchMode, submitOnToggle, patternType, setPatternType]
    )

    // This is hacky and is just for demo purposes. Once we have made the new
    // pattern type the default we can revert this.
    const onSelectSearchMode = useCallback(
        (mode: SearchModes): void => {
            if (mode === SearchModes.Smart) {
                onSelectSmartSearch(true)
            } else if (mode === SearchModes.PreciseNew) {
                toggleNewStandard()
            } else {
                onSelectSmartSearch(false)
            }
        },
        [onSelectSmartSearch, toggleNewStandard]
    )

    return (
        <div className={classNames(className, styles.toggleContainer)}>
            <>
                <QueryInputToggle
                    title="Case sensitivity"
                    isActive={caseSensitive}
                    onToggle={toggleCaseSensitivity}
                    iconSvgPath={mdiFormatLetterCase}
                    interactive={props.interactive}
                    className={`test-case-sensitivity-toggle ${styles.caseSensitivityToggle}`}
                    disableOn={[
                        {
                            condition: findFilter(navbarSearchQuery, 'case', FilterKind.Subexpression) !== undefined,
                            reason: 'Query already contains one or more case subexpressions',
                        },
                        {
                            condition:
                                findFilter(navbarSearchQuery, 'patterntype', FilterKind.Subexpression) !== undefined,
                            reason: 'Query contains one or more patterntype subexpressions, cannot apply global case-sensitivity',
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
                    iconSvgPath={mdiRegex}
                    interactive={props.interactive}
                    className={`test-regexp-toggle ${styles.regularExpressionToggle}`}
                    disableOn={[
                        {
                            condition:
                                findFilter(navbarSearchQuery, 'patterntype', FilterKind.Subexpression) !== undefined,
                            reason: 'Query already contains one or more patterntype subexpressions',
                        },
                    ]}
                />
                <>
                    {!structuralSearchDisabled && (
                        <QueryInputToggle
                            title="Structural search"
                            className={`test-structural-search-toggle ${styles.structuralSearchToggle}`}
                            isActive={patternType === SearchPatternType.structural}
                            onToggle={toggleStructuralSearch}
                            iconSvgPath={mdiCodeBrackets}
                            interactive={props.interactive}
                            disableOn={[
                                {
                                    condition:
                                        findFilter(navbarSearchQuery, 'patterntype', FilterKind.Subexpression) !==
                                        undefined,
                                    reason: 'Query already contains one or more patterntype subexpressions',
                                },
                            ]}
                        />
                    )}
                </>
                {showSmartSearchButton && <div className={styles.separator} />}
                {showSmartSearchButton &&
                    (showExtendedPicker ? (
                        <SmartSearchToggleExtended
                            className="test-smart-search-toggle"
                            mode={
                                patternType === SearchPatternType.newStandardRC1
                                    ? SearchModes.PreciseNew
                                    : searchMode === SearchMode.SmartSearch
                                    ? SearchModes.Smart
                                    : SearchModes.Precise
                            }
                            onSelect={onSelectSearchMode}
                            interactive={props.interactive}
                        />
                    ) : (
                        <SmartSearchToggle
                            className="test-smart-search-toggle"
                            isActive={searchMode === SearchMode.SmartSearch}
                            onSelect={onSelectSmartSearch}
                            interactive={props.interactive}
                        />
                    ))}
            </>
        </div>
    )
}
