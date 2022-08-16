// This file is a fork from Toggles.tsx and contains JetBrains specific UI changes
/* eslint-disable no-restricted-imports */

import React, { useCallback, useMemo } from 'react'

import classNames from 'classnames'
import CodeBracketsIcon from 'mdi-react/CodeBracketsIcon'
import FormatLetterCaseIcon from 'mdi-react/FormatLetterCaseIcon'
import LightningBoltIcon from 'mdi-react/LightningBoltIcon'
import RegexIcon from 'mdi-react/RegexIcon'

import { isErrorLike } from '@sourcegraph/common'
import {
    CaseSensitivityProps,
    SearchContextProps,
    SearchPatternTypeMutationProps,
    SearchPatternTypeProps,
    SubmitSearchProps,
} from '@sourcegraph/search'
import { QueryInputToggle } from '@sourcegraph/search-ui/src/input/toggles/QueryInputToggle'
import { SearchPatternType } from '@sourcegraph/shared/src/schema'
import { FilterKind, findFilter } from '@sourcegraph/shared/src/search/query/query'
import { appendContextFilter } from '@sourcegraph/shared/src/search/query/transformer'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { Button } from '@sourcegraph/wildcard'

import styles from './JetBrainsToggles.module.scss'

export interface JetBrainsTogglesProps
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
    clearSearch: () => void
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
export const JetBrainsToggles: React.FunctionComponent<React.PropsWithChildren<JetBrainsTogglesProps>> = (
    props: JetBrainsTogglesProps
) => {
    const {
        navbarSearchQuery,
        patternType,
        setPatternType,
        caseSensitive,
        setCaseSensitivity,
        settingsCascade,
        className,
        submitSearch,
        structuralSearchDisabled,
        clearSearch,
    } = props

    const defaultPatternTypeValue = useMemo(
        () =>
            settingsCascade.final &&
            !isErrorLike(settingsCascade.final) &&
            (settingsCascade.final['search.defaultPatternType'] as SearchPatternType),
        [settingsCascade.final]
    )

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
        const defaultPatternType = defaultPatternTypeValue || SearchPatternType.standard

        const newPatternType: SearchPatternType =
            patternType !== SearchPatternType.structural ? SearchPatternType.structural : defaultPatternType

        setPatternType(newPatternType)
        submitOnToggle({ newPatternType })
    }, [defaultPatternTypeValue, patternType, setPatternType, submitOnToggle])

    const toggleExpertMode = useCallback((): void => {
        const newPatternType =
            patternType === SearchPatternType.lucky ? SearchPatternType.standard : SearchPatternType.lucky

        setPatternType(newPatternType)
        submitOnToggle({ newPatternType })
    }, [patternType, setPatternType, submitOnToggle])

    const luckySearchEnabled = defaultPatternTypeValue === SearchPatternType.lucky

    return (
        <div className={classNames(className, styles.toggleContainer)}>
            {navbarSearchQuery !== '' && (
                <Button
                    variant="icon"
                    className={classNames(props.className, styles.cancelButton)}
                    onClick={clearSearch}
                >
                    <span aria-hidden="true">&times;</span>
                </Button>
            )}
            <div className={styles.separator} />
            {patternType === SearchPatternType.lucky ? (
                <>
                    <QueryInputToggle
                        title="Expert mode"
                        isActive={false}
                        onToggle={toggleExpertMode}
                        icon={LightningBoltIcon}
                        interactive={props.interactive}
                        className={classNames(styles.toggle, 'test-expert-mode-toggle')}
                        activeClassName="test-expert-mode-toggle--active"
                        disableOn={[]}
                    />
                </>
            ) : (
                <>
                    <QueryInputToggle
                        title="Case sensitivity"
                        isActive={caseSensitive}
                        onToggle={toggleCaseSensitivity}
                        icon={FormatLetterCaseIcon}
                        interactive={props.interactive}
                        className={classNames(styles.toggle, 'test-case-sensitivity-toggle')}
                        activeClassName="test-case-sensitivity-toggle--active"
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
                        icon={RegexIcon}
                        interactive={props.interactive}
                        className={classNames(styles.toggle, 'test-regexp-toggle')}
                        activeClassName="test-regexp-toggle--active"
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
                            className={classNames(styles.toggle, 'test-structural-search-toggle')}
                            activeClassName="test-structural-search-toggle--active"
                            isActive={patternType === SearchPatternType.structural}
                            onToggle={toggleStructuralSearch}
                            icon={CodeBracketsIcon}
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
                    {luckySearchEnabled && (
                        <QueryInputToggle
                            title="Expert mode"
                            isActive={true}
                            onToggle={toggleExpertMode}
                            icon={LightningBoltIcon}
                            interactive={props.interactive}
                            className="test-expert-mode-toggle"
                            activeClassName="test-expert-mode-toggle--active"
                            disableOn={[]}
                        />
                    )}
                </>
            )}
        </div>
    )
}
