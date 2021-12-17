import classNames from 'classnames'
import Check from 'mdi-react/CheckIcon'
import CloseIcon from 'mdi-react/CloseIcon'
import RadioboxBlankIcon from 'mdi-react/RadioboxBlankIcon'
import React from 'react'

import styles from './SearchQueryChecks.module.scss'

interface SearchQueryChecksProps {
    checks: {
        isValidOperator: true | false | undefined
        isValidPatternType: true | false | undefined
        isNotRepo: true | false | undefined
        isNotCommitOrDiff: true | false | undefined
    }
}

const CheckListItem: React.FunctionComponent<{ valid: true | false | undefined }> = ({ children, valid }) => {
    if (valid === true) {
        return (
            <>
                <Check className={classNames(styles.icon, 'text-success icon-inline')} />
                <span className={classNames(styles.valid, 'text-muted')}>{children}</span>
            </>
        )
    }

    if (valid === false) {
        return (
            <>
                <CloseIcon className={classNames(styles.icon, 'text-danger icon-inline')} />
                <span className="text-muted">{children}</span>
            </>
        )
    }

    return (
        <>
            <RadioboxBlankIcon className={classNames(styles.icon, styles.smaller, 'icon-inline')} />{' '}
            <span className="text-muted">{children}</span>
        </>
    )
}

export const SearchQueryChecks: React.FunctionComponent<SearchQueryChecksProps> = ({ checks }) => (
    <div className={classNames(styles.checksWrapper)}>
        <ul className={classNames(styles.checks)}>
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
