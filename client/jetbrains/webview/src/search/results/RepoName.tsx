import classNames from 'classnames'

import { Tooltip, useIsTruncated } from '@sourcegraph/wildcard'

import styles from './RepoName.module.scss'

interface RepoNameProps {
    repoName: string
    suffix?: string
}

export const RepoName: React.FunctionComponent<React.PropsWithChildren<RepoNameProps>> = ({ repoName, suffix }) => {
    /**
     * Use the custom hook useIsTruncated to check for an overflow: ellipsis is activated for the element
     * We want to do it on mouse enter as browser window size might change after the element has been
     * loaded initially
     */
    const [titleReference, truncated, checkTruncation] = useIsTruncated()

    return (
        <div className={classNames('w-100', styles.forceSiblingToW100)}>
            <Tooltip content={truncated && suffix ? suffix : null} placement="bottom">
                <div ref={titleReference} onMouseEnter={checkTruncation} className="text-truncate">
                    {repoName}
                    {suffix ? ` › ${suffix}` : null}
                </div>
            </Tooltip>
        </div>
    )
}
