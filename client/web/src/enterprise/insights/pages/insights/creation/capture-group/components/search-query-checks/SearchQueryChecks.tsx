import classNames from 'classnames'
import Check from 'mdi-react/CheckIcon'
import CloseIcon from 'mdi-react/CloseIcon'
import React from 'react'

import styles from './SearchQueryChecks.module.scss'

interface SearchQueryChecksProps {
    checks: {
        isValidRegex: boolean
        isValidOperator: boolean
        isValidPatternType: boolean
        isNotRepo: boolean
        isNotCommitOrDiff: boolean
    }
}

const CheckListItem: React.FunctionComponent<{ valid?: boolean }> = ({ children, valid }) =>
    valid ? (
        <span>
            <Check size={16} className="text-success icon-inline" style={{ top: '3px' }} /> {children}
        </span>
    ) : (
        <span className="text-dark">
            <CloseIcon size={16} className="text-danger icon-inline" style={{ top: '3px' }} /> {children}
        </span>
    )

export const SearchQueryChecks: React.FunctionComponent<SearchQueryChecksProps> = ({ checks }) => (
    <div className={classNames(styles.checks)}>
        <ul className={classNames('text-muted', styles.check)}>
            <li>
                <CheckListItem valid={checks.isValidRegex}>
                    Contains a properly formatted regular expression
                </CheckListItem>
            </li>
            <li>
                <CheckListItem valid={checks.isValidOperator}>
                    Does not contain boolean operator <code>AND</code> and <code>OR</code> (regular expression boolean
                    operators can still be used)
                </CheckListItem>
            </li>
            <li>
                <CheckListItem valid={checks.isValidPatternType}>
                    Does not contain <code>patternType:literal</code> and <code>patternType:structural</code>
                </CheckListItem>
            </li>
            <li>
                <CheckListItem valid={checks.isNotRepo}>
                    Does not contain <code>repo:</code> filter
                </CheckListItem>
            </li>
            <li>
                <CheckListItem valid={checks.isNotCommitOrDiff}>
                    Does not contain <code>commit</code> or <code>diff</code> search
                </CheckListItem>
            </li>
        </ul>

        <p className="mt-1 text-muted">
            Tip: use <code>archived:no</code> or <code>fork:no</code> to exclude results from archived or forked
            repositories.
        </p>
    </div>
)
