import classNames from 'classnames'
import BookOpenVariantIcon from 'mdi-react/BookOpenVariantIcon'
import React from 'react'

import { DismissibleAlert } from '@sourcegraph/web/src/components/DismissibleAlert'

import styles from './DocumentationWelcomeAlert.module.scss'

export const DocumentationWelcomeAlert: React.FunctionComponent = () => (
    <DismissibleAlert
        className={classNames('mt-3', styles.documentationWelcomeAlert)}
        partialStorageKey="apidocs-welcome"
    >
        <div className="card">
            <div className="card-body p-3">
                <h1>
                    <BookOpenVariantIcon className="icon-inline mr-2" />
                    API docs, for your code
                </h1>
                <ul className="mb-0 pl-3">
                    <li>Use the navbar on the left to navigate all the API documentation for this repository.</li>
                    <li>Only the Go programming language is supported at this time.</li>
                    <li>
                        <a
                            // eslint-disable-next-line react/jsx-no-target-blank
                            target="_blank"
                            rel="noopener"
                            href="https://docs.sourcegraph.com/code_intelligence/apidocs"
                        >
                            Learn more
                        </a>
                    </li>
                </ul>
            </div>
        </div>
    </DismissibleAlert>
)
