import React from 'react'

import classNames from 'classnames'
import Check from 'mdi-react/CheckIcon'
import CloseIcon from 'mdi-react/CloseIcon'
import RadioboxBlankIcon from 'mdi-react/RadioboxBlankIcon'

import { Icon, Code } from '@sourcegraph/wildcard'

import styles from './SearchQueryChecks.module.scss'

interface SearchQueryChecksProps {
    checks: {
        isValidOperator: true | false | undefined
        isValidPatternType: true | false | undefined
        isNotRepo: true | false | undefined
        isNotCommitOrDiff: true | false | undefined
        isNoNewLines: true | false | undefined
    }
}

const CheckListItem: React.FunctionComponent<React.PropsWithChildren<{ valid: true | false | undefined }>> = ({
    children,
    valid,
}) => {
    if (valid === true) {
        return (
            <>
                <Icon aria-hidden={true} className={classNames(styles.icon, 'text-success')} as={Check} />
                <span className={classNames(styles.valid, 'text-muted')}>{children}</span>
            </>
        )
    }

    if (valid === false) {
        return (
            <>
                <Icon aria-hidden={true} className={classNames(styles.icon, 'text-danger')} as={CloseIcon} />
                <span className="text-muted">{children}</span>
            </>
        )
    }

    return (
        <>
            <Icon aria-hidden={true} className={classNames(styles.icon, styles.smaller)} as={RadioboxBlankIcon} />{' '}
            <span className="text-muted">{children}</span>
        </>
    )
}

export const SearchQueryChecks: React.FunctionComponent<React.PropsWithChildren<SearchQueryChecksProps>> = ({
    checks,
}) => (
    <div className={classNames(styles.checksWrapper)}>
        <ul className={classNames(styles.checks)}>
            <li>
                <CheckListItem valid={checks.isNoNewLines}>
                    Does not contain a match over more than a single line.
                </CheckListItem>
            </li>
            <li>
                <CheckListItem valid={checks.isValidOperator}>
                    Does not contain boolean operators <Code>AND</Code>, <Code>OR</Code>, and <Code>NOT</Code> (regular
                    expression boolean operators can still be used)
                </CheckListItem>
            </li>
            <li>
                <CheckListItem valid={checks.isValidPatternType}>
                    Does not contain <Code>patternType:literal</Code> and <Code>patternType:structural</Code>
                </CheckListItem>
            </li>
            <li>
                <CheckListItem valid={checks.isNotRepo}>
                    Does not contain <Code>repo:</Code> filter
                </CheckListItem>
            </li>
            <li>
                <CheckListItem valid={checks.isNotCommitOrDiff}>
                    Does not contain <Code>commit</Code> or <Code>diff</Code> search
                </CheckListItem>
            </li>
        </ul>
    </div>
)
