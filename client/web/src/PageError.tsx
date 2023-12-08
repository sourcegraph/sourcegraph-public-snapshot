import ServerIcon from 'mdi-react/ServerIcon'

import { FeedbackText } from '@sourcegraph/wildcard'

import { HeroPage } from './components/HeroPage'

import styles from './PageError.module.scss'

// CHANGE TO TRIGGER BUILD

interface Props {
    pageError: PageError
}
export const PageError: React.FC<Props> = ({ pageError }) => {
    const statusCode = pageError.statusCode
    const statusText = pageError.statusText
    const errorMessage = pageError.error
    const errorID = pageError.errorID

    let subtitle: JSX.Element | undefined
    if (errorID) {
        subtitle = <FeedbackText headerText="Sorry, there's been a problem." />
    }
    if (errorMessage) {
        subtitle = (
            <div className={styles.error}>
                {subtitle}
                {subtitle && <hr className="my-3" />}
                <pre>{errorMessage}</pre>
            </div>
        )
    } else {
        subtitle = <div className={styles.error}>{subtitle}</div>
    }

    return <HeroPage icon={ServerIcon} title={`${statusCode}: ${statusText}`} subtitle={subtitle} />
}
