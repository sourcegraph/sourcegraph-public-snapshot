import * as React from 'react'
import * as H from 'history'
import RegexIcon from 'mdi-react/RegexIcon'
import FormatLetterCaseIcon from 'mdi-react/FormatLetterCaseIcon'
import { PatternTypeProps, CaseSensitivityProps } from '../..'
import { FiltersToTypeAndValue } from '../../../../../shared/src/search/interactive/util'
import { SettingsCascadeProps } from '../../../../../shared/src/settings/settings'
import { SearchPatternType } from '../../../../../shared/src/graphql/schema'
import { isEmpty } from 'lodash'
import { submitSearch } from '../../helpers'
import { QueryInputToggle } from './QueryInputToggle'
import { isErrorLike } from '../../../../../shared/src/util/errors'
import CodeBracketsIcon from 'mdi-react/CodeBracketsIcon'
import { ActivationProps } from '../../../../../shared/src/components/activation/Activation'

export interface TogglesProps extends PatternTypeProps, CaseSensitivityProps, SettingsCascadeProps {
    navbarSearchQuery: string
    history: H.History
    location: H.Location
    filtersInQuery?: FiltersToTypeAndValue
    hasGlobalQueryBehavior?: boolean
}

interface SubmitSearchArgs {
    history: H.History
    navbarQuery: string
    source: 'home' | 'nav' | 'repo' | 'tree' | 'filter' | 'type'
    patternType: SearchPatternType
    caseSensitive: boolean
    activation?: ActivationProps['activation']
    filtersQuery?: FiltersToTypeAndValue
}

/**
 * The toggles displayed in the query input.
 */
export const Toggles: React.FunctionComponent<TogglesProps> = (props: TogglesProps) => {
    const structuralSearchDisabled =
        window.context &&
        window.context.experimentalFeatures &&
        window.context.experimentalFeatures.structuralSearch === 'disabled'

    const submitOnToggle = (args: SubmitSearchArgs): void => {
        const searchQueryNotEmpty =
            props.navbarSearchQuery !== '' || (props.filtersInQuery && !isEmpty(props.filtersInQuery))
        const shouldSubmitSearch = props.hasGlobalQueryBehavior && searchQueryNotEmpty

        if (shouldSubmitSearch) {
            // Only submit search on toggle when the query input has global behavior (i.e. it's on the main search page
            // or global navbar). Non-global inputs don't have the canonical query and need more context, making
            // submit on-toggle undesirable. Also, only submit on toggle only when the query is non-empty.
            submitSearch(
                args.history,
                args.navbarQuery,
                args.source,
                args.patternType,
                args.caseSensitive,
                args.activation,
                args.filtersQuery
            )
        }
    }

    const toggleCaseSensitivity = (): void => {
        if (props.patternType === SearchPatternType.structural) {
            return
        }

        const newCaseSensitivity = !props.caseSensitive
        props.setCaseSensitivity(newCaseSensitivity)

        submitOnToggle({
            history: props.history,
            navbarQuery: props.navbarSearchQuery,
            source: 'filter',
            patternType: props.patternType,
            caseSensitive: newCaseSensitivity,
            activation: undefined,
            filtersQuery: props.filtersInQuery,
        })
    }

    const toggleRegexp = (): void => {
        if (props.patternType === SearchPatternType.structural) {
            return
        }

        const newPatternType =
            props.patternType !== SearchPatternType.regexp ? SearchPatternType.regexp : SearchPatternType.literal

        props.setPatternType(newPatternType)

        submitOnToggle({
            history: props.history,
            navbarQuery: props.navbarSearchQuery,
            source: 'filter',
            patternType: newPatternType,
            caseSensitive: props.caseSensitive,
            activation: undefined,
            filtersQuery: props.filtersInQuery,
        })
    }

    const toggleStructuralSearch = (): void => {
        const cascadePatternTypeValue =
            props.settingsCascade.final &&
            !isErrorLike(props.settingsCascade.final) &&
            props.settingsCascade.final['search.defaultPatternType']

        const defaultPatternType = cascadePatternTypeValue || 'literal'

        const newPatternType =
            props.patternType !== SearchPatternType.structural ? SearchPatternType.structural : defaultPatternType

        props.setPatternType(newPatternType)

        submitOnToggle({
            history: props.history,
            navbarQuery: props.navbarSearchQuery,
            source: 'filter',
            patternType: newPatternType,
            caseSensitive: props.caseSensitive,
            activation: undefined,
            filtersQuery: props.filtersInQuery,
        })
    }

    return (
        <div className="query-input2__toggle-container">
            <QueryInputToggle
                {...props}
                title="Case sensitivity"
                isActive={props.caseSensitive}
                onToggle={toggleCaseSensitivity}
                icon={FormatLetterCaseIcon}
                className="e2e-case-sensitivity-toggle"
                activeClassName="e2e-case-sensitivity-toggle--active"
                disabledCondition={props.patternType === SearchPatternType.structural}
                disabledMessage="Structural search is always case sensitive"
            />
            <QueryInputToggle
                {...props}
                title="Regular expression"
                isActive={props.patternType === SearchPatternType.regexp}
                onToggle={toggleRegexp}
                icon={RegexIcon}
                className="e2e-regexp-toggle"
                activeClassName="e2e-regexp-toggle--active"
                disabledCondition={props.patternType === SearchPatternType.structural}
                disabledMessage="Structural search uses Comby syntax"
            />
            {!structuralSearchDisabled && (
                <QueryInputToggle
                    {...props}
                    title="Structural search"
                    className="e2e-structural-search-toggle"
                    activeClassName="e2e-structural-search-toggle--active"
                    isActive={props.patternType === SearchPatternType.structural}
                    onToggle={toggleStructuralSearch}
                    icon={CodeBracketsIcon}
                />
            )}
        </div>
    )
}
