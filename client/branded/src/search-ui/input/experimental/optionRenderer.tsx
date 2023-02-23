import classnames from 'classnames'

import { HighlightedLabel } from './Suggestions'
import { Option } from './suggestionsExtension'
import { SyntaxHighlightedSearchQuery } from './SyntaxHighlightedSearchQuery'

import styles from './Suggestions.module.scss'

const FilterOption: React.FunctionComponent<{ option: Option }> = ({ option }) => (
    <span className={classnames(styles.filterOption, styles.filterField)}>
        {option.matches ? <HighlightedLabel label={option.label} matches={option.matches} /> : option.label}
        <span className={styles.separator}>:</span>
    </span>
)

const FilterValueOption: React.FunctionComponent<{ option: Option }> = ({ option }) => {
    const label = option.label
    const separatorIndex = label.indexOf(':')
    const field = label.slice(0, separatorIndex)
    const value = label.slice(separatorIndex + 1)

    return (
        <span className={styles.filterOption}>
            <span className={styles.filterField}>
                {field}
                <span className={styles.separator}>:</span>
            </span>
            {option.matches ? <HighlightedLabel label={value} matches={option.matches} /> : option.label}
        </span>
    )
}

const QueryOption: React.FunctionComponent<{ option: Option }> = ({ option }) => (
    <SyntaxHighlightedSearchQuery query={option.label} matches={option.matches} />
)

// Custom renderer for filter suggestions
export const filterRenderer = (option: Option): React.ReactElement => <FilterOption option={option} />
export const filterValueRenderer = (option: Option): React.ReactElement => <FilterValueOption option={option} />
// Custom renderer for (the current) query suggestions
export const queryRenderer = (option: Option): React.ReactElement => <QueryOption option={option} />
export const submitQueryInfo = (): React.ReactElement => (
    <>
        Press <kbd>Return</kbd> to submit your query.
    </>
)
