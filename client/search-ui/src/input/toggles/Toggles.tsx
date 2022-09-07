import React, { useCallback, useMemo } from 'react'

import { mdiCodeBrackets, mdiFormatLetterCase, mdiRegex } from '@mdi/js'
import classNames from 'classnames'

import { isErrorLike, isMacPlatform } from '@sourcegraph/common'
import {
    SearchPatternTypeProps,
    CaseSensitivityProps,
    SearchContextProps,
    SearchPatternTypeMutationProps,
    SubmitSearchProps,
} from '@sourcegraph/search'
import { SearchPatternType } from '@sourcegraph/shared/src/schema'
import { findFilter, FilterKind } from '@sourcegraph/shared/src/search/query/query'
import { appendContextFilter } from '@sourcegraph/shared/src/search/query/transformer'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'

import { CopyQueryButton } from './CopyQueryButton'
import { QueryInputToggle } from './QueryInputToggle'
import { SmartSearchToggle } from './SmartSearchToggle'

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

    const defaultPatternTypeValue = useMemo(
        () =>
            (settingsCascade.final &&
                !isErrorLike(settingsCascade.final) &&
                (settingsCascade.final['search.defaultPatternType'] as SearchPatternType)) ||
            SearchPatternType.standard,
        [settingsCascade.final]
    )

    const showSmartSearch = useMemo(() => defaultPatternTypeValue === SearchPatternType.lucky, [
        defaultPatternTypeValue,
    ])

    const submitOnToggle = useCallback(
        (
            args: { newPatternType: SearchPatternType } | { newCaseSensitivity: boolean } | { newPowerUser: boolean }
        ): void => {
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
            patternType !== SearchPatternType.regexp ? SearchPatternType.regexp : SearchPatternType.standard

        setPatternType(newPatternType)
        submitOnToggle({ newPatternType })
    }, [patternType, setPatternType, submitOnToggle])

    const toggleStructuralSearch = useCallback((): void => {
        const newPatternType: SearchPatternType =
            patternType !== SearchPatternType.structural ? SearchPatternType.structural : SearchPatternType.standard

        setPatternType(newPatternType)
        submitOnToggle({ newPatternType })
    }, [patternType, setPatternType, submitOnToggle])

    const onSelectSmartSearch = useCallback(
        (enabled: boolean): void => {
            const newPatternType: SearchPatternType = enabled ? SearchPatternType.lucky : SearchPatternType.standard

            setPatternType(newPatternType)
            submitOnToggle({ newPatternType })
        },
        [setPatternType, submitOnToggle]
    )

    const fullQuery = getFullQuery(navbarSearchQuery, selectedSearchContextSpec || '', caseSensitive, patternType)

    return (
        <div className={classNames(className, styles.toggleContainer)}>
            <>
                {/* Hide the other toggles if lucky search is enabled */}
                {(!showSmartSearch || patternType !== SearchPatternType.lucky) && (
                    <>
                        <QueryInputToggle
                            title="Case sensitivity"
                            isActive={caseSensitive}
                            onToggle={toggleCaseSensitivity}
                            iconSvgPath={mdiFormatLetterCase}
                            interactive={props.interactive}
                            className="test-case-sensitivity-toggle"
                            disableOn={[
                                {
                                    condition:
                                        findFilter(navbarSearchQuery, 'case', FilterKind.Subexpression) !== undefined,
                                    reason: 'Query already contains one or more case subexpressions',
                                },
                                {
                                    condition:
                                        findFilter(navbarSearchQuery, 'patterntype', FilterKind.Subexpression) !==
                                        undefined,
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
                            iconSvgPath={mdiRegex}
                            interactive={props.interactive}
                            className="test-regexp-toggle"
                            disableOn={[
                                {
                                    condition:
                                        findFilter(navbarSearchQuery, 'patterntype', FilterKind.Subexpression) !==
                                        undefined,
                                    reason: 'Query already contains one or more patterntype subexpressions',
                                },
                            ]}
                        />
                        {!structuralSearchDisabled && (
                            <QueryInputToggle
                                title="Structural search"
                                className="test-structural-search-toggle"
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
                        {(showSmartSearch || showCopyQueryButton) && <div className={styles.separator} />}
                    </>
                )}
                {showSmartSearch && (
                    <SmartSearchToggle
                        className="test-smart-search-toggle"
                        isActive={patternType === SearchPatternType.lucky}
                        onSelect={onSelectSmartSearch}
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
                {showCopyQueryButton && (
                    <CopyQueryButton
                        fullQuery={fullQuery}
                        isMacPlatform={isMacPlatform()}
                        className={classNames(styles.toggle, styles.copyQueryButton)}
                    />
                )}
            </>
        </div>
    )
}
