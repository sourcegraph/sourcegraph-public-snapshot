import classNames from 'classnames'
import BookOpenBlankVariantIcon from 'mdi-react/BookOpenBlankVariantIcon'
import React from 'react'

import { DismissibleAlert } from '@sourcegraph/web/src/components/DismissibleAlert'
import { CardBody, Card, Link } from '@sourcegraph/wildcard'

import styles from './DocumentationWelcomeAlert.module.scss'

export const DocumentationWelcomeAlert: React.FunctionComponent = () => (
    <DismissibleAlert
        className={classNames('mt-3', styles.documentationWelcomeAlert)}
        partialStorageKey="apidocs-welcome"
    >
        <Card>
            <CardBody>
                <h1>
                    <BookOpenBlankVariantIcon className="icon-inline mr-2" />
                    API docs, for your code
                </h1>
                <ul className="mb-0 pl-3">
                    <li>Use the navbar on the left to navigate all the API documentation for this repository.</li>
                    <li>Only the Go programming language is supported at this time.</li>
                    <li>
                        <Link
                            target="_blank"
                            rel="noopener"
                            to="https://docs.sourcegraph.com/code_intelligence/apidocs"
                        >
                            Learn more
                        </Link>
                    </li>
                </ul>
            </CardBody>
        </Card>
    </DismissibleAlert>
)
