import { SyntaxHighlightedSearchQuery } from '../../components'

import { HighlightedLabel } from './Suggestions'
import { Option } from './suggestionsExtension'

import styles from './Suggestions.module.scss'

const FilterOption: React.FunctionComponent<{ option: Option }> = ({ option }) => (
    <span className={styles.filterOption}>
        {option.matches ? <HighlightedLabel label={option.label} matches={option.matches} /> : option.label}
        <span className={styles.separator}>:</span>
    </span>
)

const QueryOption: React.FunctionComponent<{ option: Option }> = ({ option }) => (
    <SyntaxHighlightedSearchQuery query={option.label} />
)

// Custom renderer for filter suggestions
export const filterRenderer = (option: Option): React.ReactElement => <FilterOption option={option} />
// Custom renderer for (the current) query suggestions
export const queryRenderer = (option: Option): React.ReactElement => <QueryOption option={option} />
export const submitQueryInfo = (): React.ReactElement => (
    <>
        Press <kbd>Return</kbd> to submit your query.
    </>
)
