import { usePrettifyEditors, useHistoryContext } from '@graphiql/react'
import graphiql from 'graphiql'

import { Button, Alert, ButtonLink } from '@sourcegraph/wildcard'

import styles from './ApiConsoleToolbar.module.scss'

export const ApiConsoleToolbar: React.FunctionComponent = () => {
    const prettify = usePrettifyEditors()
    const historyContext = useHistoryContext()

    return (
        <graphiql.Toolbar>
            <div className="d-flex align-items-center">
                <Button
                    variant="secondary"
                    title="Prettify Query (Shift-Ctrl-P)"
                    onClick={() => prettify()}
                    className={styles.toolbarButton}
                >
                    Prettify
                </Button>
                <Button
                    variant="secondary"
                    title="Show History"
                    onClick={() => historyContext?.toggle()}
                    className={styles.toolbarButton}
                >
                    History
                </Button>
                <ButtonLink to="/help/api/graphql" variant="link">
                    Docs
                </ButtonLink>
                <Alert variant="warning" className="py-1 mb-0 ml-2 text-nowrap">
                    <small>
                        The API console uses <strong>real production data.</strong>
                    </small>
                </Alert>
            </div>
        </graphiql.Toolbar>
    )
}

ApiConsoleToolbar.displayName = 'GraphiQLToolbar'
