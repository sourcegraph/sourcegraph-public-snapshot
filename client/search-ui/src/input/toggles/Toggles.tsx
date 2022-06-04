import React, { useCallback, useMemo } from 'react'

import classNames from 'classnames'
import CodeBracketsIcon from 'mdi-react/CodeBracketsIcon'
import FormatLetterCaseIcon from 'mdi-react/FormatLetterCaseIcon'
import LightningBoltIcon from 'mdi-react/LightningBoltIcon'
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
        ...otherProps
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
            patternType !== SearchPatternType.regexp ? SearchPatternType.regexp : SearchPatternType.literal

        setPatternType(newPatternType)
        submitOnToggle({ newPatternType })
    }, [patternType, setPatternType, submitOnToggle])

    const toggleStructuralSearch = useCallback((): void => {
        const defaultPatternType = defaultPatternTypeValue || SearchPatternType.literal

        const newPatternType: SearchPatternType =
            patternType !== SearchPatternType.structural ? SearchPatternType.structural : defaultPatternType

        setPatternType(newPatternType)
        submitOnToggle({ newPatternType })
    }, [defaultPatternTypeValue, patternType, setPatternType, submitOnToggle])

    const toggleExpertMode = useCallback((): void => {
        const newPatternType =
            patternType === SearchPatternType.lucky ? SearchPatternType.literal : SearchPatternType.lucky

        setPatternType(newPatternType)
        submitOnToggle({ newPatternType })
    }, [patternType, setPatternType, submitOnToggle])

    const luckySearchEnabled = defaultPatternTypeValue === SearchPatternType.lucky

    const fullQuery = getFullQuery(navbarSearchQuery, selectedSearchContextSpec || '', caseSensitive, patternType)

    const copyQueryButton = showCopyQueryButton && (
        <>
            <div className={styles.separator} />
            <CopyQueryButton
                fullQuery={fullQuery}
                keyboardShortcutForFullCopy={KEYBOARD_SHORTCUT_COPY_FULL_QUERY}
                isMacPlatform={isMacPlatform()}
                className={classNames(styles.toggle, styles.copyQueryButton)}
            />
        </>
    )

    return (
        <div className={classNames(className, styles.toggleContainer)} {...otherProps}>
            {patternType === SearchPatternType.lucky ? (
                <>
                    <QueryInputToggle
                        title="Expert mode"
                        isActive={false}
                        onToggle={toggleExpertMode}
                        icon={LightningBoltIcon}
                        interactive={props.interactive}
                        className="test-expert-mode-toggle"
                        activeClassName="test-expert-mode-toggle--active"
                        disableOn={[]}
                    />
                    {copyQueryButton}
                </>
            ) : (
                <>
                    <QueryInputToggle
                        title="Case sensitivity"
                        isActive={caseSensitive}
                        onToggle={toggleCaseSensitivity}
                        icon={FormatLetterCaseIcon}
                        interactive={props.interactive}
                        className="test-case-sensitivity-toggle"
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
                        className="test-regexp-toggle"
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
                            className="test-structural-search-toggle"
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
                    {copyQueryButton}
                </>
            )}
        </div>
    )
}
