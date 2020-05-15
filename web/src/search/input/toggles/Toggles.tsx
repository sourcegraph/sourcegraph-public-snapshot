import * as React from 'react'
import * as H from 'history'
import RegexIcon from 'mdi-react/RegexIcon'
import classNames from 'classnames'
import FormatLetterCaseIcon from 'mdi-react/FormatLetterCaseIcon'
import { PatternTypeProps, CaseSensitivityProps, InteractiveSearchProps, CopyQueryButtonProps } from '../..'
import { SettingsCascadeProps } from '../../../../../shared/src/settings/settings'
import { SearchPatternType } from '../../../../../shared/src/graphql/schema'
import { isEmpty } from 'lodash'
import { submitSearch } from '../../helpers'
import { QueryInputToggle } from './QueryInputToggle'
import { isErrorLike } from '../../../../../shared/src/util/errors'
import CodeBracketsIcon from 'mdi-react/CodeBracketsIcon'
import { generateFiltersQuery } from '../../../../../shared/src/util/url'
import { CopyQueryButton } from './CopyQueryButton'
import { VersionContextProps } from '../../../../../shared/src/search/util'

export interface TogglesProps
    extends PatternTypeProps,
        CaseSensitivityProps,
        SettingsCascadeProps,
        CopyQueryButtonProps,
        Partial<Pick<InteractiveSearchProps, 'filtersInQuery'>>,
        VersionContextProps {
    navbarSearchQuery: string
    history: H.History
    location: H.Location
    hasGlobalQueryBehavior?: boolean
    className?: string
}

/**
 * The toggles displayed in the query input.
 */
export const Toggles: React.FunctionComponent<TogglesProps> = (props: TogglesProps) => {
    const structuralSearchDisabled =
        window.context &&
        window.context.experimentalFeatures &&
        window.context.experimentalFeatures.structuralSearch === 'disabled'

    const submitOnToggle = (args: { newPatternType: SearchPatternType } | { newCaseSensitivity: boolean }): void => {
        const { history, navbarSearchQuery, filtersInQuery, versionContext } = props
        const searchQueryNotEmpty = navbarSearchQuery !== '' || (filtersInQuery && !isEmpty(filtersInQuery))
        const shouldSubmitSearch = props.hasGlobalQueryBehavior && searchQueryNotEmpty
        const activation = undefined
        const source = 'filter'
        const patternType = 'newPatternType' in args ? args.newPatternType : props.patternType
        const caseSensitive = 'newCaseSensitivity' in args ? args.newCaseSensitivity : props.caseSensitive
        if (shouldSubmitSearch) {
            // Only submit search on toggle when the query input has global behavior (i.e. it's on the main search page
            // or global navbar). Non-global inputs don't have the canonical query and need more context, making
            // submit on-toggle undesirable. Also, only submit on toggle only when the query is non-empty.
            submitSearch({
                history,
                query: navbarSearchQuery,
                source,
                patternType,
                caseSensitive,
                versionContext,
                activation,
                filtersInQuery,
            })
        }
    }

    const toggleCaseSensitivity = (): void => {
        if (props.patternType === SearchPatternType.structural) {
            return
        }
        const newCaseSensitivity = !props.caseSensitive
        props.setCaseSensitivity(newCaseSensitivity)
        submitOnToggle({ newCaseSensitivity })
    }

    const toggleRegexp = (): void => {
        const newPatternType =
            props.patternType !== SearchPatternType.regexp ? SearchPatternType.regexp : SearchPatternType.literal

        props.setPatternType(newPatternType)
        submitOnToggle({ newPatternType })
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
        submitOnToggle({ newPatternType })
    }

    const fullQuery = [
        props.navbarSearchQuery,
        props.filtersInQuery && generateFiltersQuery(props.filtersInQuery),
        `patternType:${props.patternType}`,
        props.caseSensitive ? 'case:yes' : '',
    ]
        .filter(queryPart => !!queryPart)
        .join(' ')

    return (
        <div className={classNames('toggle-container', props.className)}>
            {props.copyQueryButton && (
                <CopyQueryButton
                    fullQuery={fullQuery}
                    className="toggle-container__toggle toggle-container__copy-query-button"
                />
            )}
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
