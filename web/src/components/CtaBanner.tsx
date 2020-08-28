import * as React from 'react'
import classNames from 'classnames'

interface Props {
    className?: string
    icon: string | React.ReactNode
}

export const CtaBanner = React.memo<Props>(({ icon, className }) => (
    <div className={classNames('web-content private-code-cta shadow d-flex flex-row card', className)}>
        <div className="private-code-cta__icon-column d-flex flex-column align-items-center">
            {typeof icon === 'string' ? <img src={icon} className="private-code-cta__logo" /> : { icon }}
        </div>
        <div>
            <h3>Search your private code</h3>
            <p>
                Set up a self-hosted Sourcegraph instance to search your private repositories on GitHub, GitLab,
                Bitbucket and local installations of Git, Perforce, Subversion and other code repositories.
            </p>
            <a
                href="https://docs.sourcegraph.com/"
                target="_blank"
                rel="noreferrer"
                className="btn btn-primary ga-cta-install-now"
            >
                Install now
            </a>
        </div>
    </div>
))
