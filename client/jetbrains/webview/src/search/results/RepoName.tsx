import classNames from 'classnames'

import { Tooltip, useIsTruncated } from '@sourcegraph/wildcard'

interface RepoNameProps {
    repoName: string
    suffix?: string
}
import styles from './RepoName.module.scss'

export const RepoName: React.FunctionComponent<React.PropsWithChildren<RepoNameProps>> = ({ repoName, suffix }) => {
    /**
     * Use the custom hook useIsTruncated to check for an overflow: ellipsis is activated for the element
     * We want to do it on mouse enter as browser window size might change after the element has been
     * loaded initially
     */
    const [titleReference, truncated, checkTruncation] = useIsTruncated()

    /**
     * Enabling a tooltip will break the React tree inside JetBrains because we polyfill
     * `window.Event` which is used internally by the tooltip positioning logic together with
     * dispatchEvent().
     *
     * c.f. https://sourcegraph.com/github.com/sourcegraph/sourcegraph/-/blob/client/jetbrains/webview/src/search/index.tsx?L47
     */
    const tooltipContent = truncated && suffix ? suffix : null

    return (
        <div className={classNames('w-100', styles.forceSiblingToW100)}>
            <Tooltip content={tooltipContent} placement="bottom">
                <div ref={titleReference} onMouseEnter={checkTruncation} className="text-truncate">
                    {repoName}
                    {suffix ? ` › ${suffix}` : null}
                </div>
            </Tooltip>
        </div>
    )
}
