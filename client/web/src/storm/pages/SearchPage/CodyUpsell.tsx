import type { FC } from 'react'

import { useIsLightTheme } from '@sourcegraph/shared/src/theme'
import { Link, Text } from '@sourcegraph/wildcard'

import { CodyLogo } from '../../../cody/components/CodyLogo'

import { MultiLineCompletion } from './MultilineCompletion'

import styles from './CodyUpsell.module.scss'

interface CodyUpsellProps {
    isSourcegraphDotCom: boolean
}

export const CodyUpsell: FC<CodyUpsellProps> = ({ isSourcegraphDotCom }) => {
    const isLightTheme = useIsLightTheme()
    // On DotCom, we want to redirect to the PLG page. On Enterprise instances, we redirect to their Cody dashboard page.
    const exploreCodyLink = isSourcegraphDotCom ? 'https://sourcegraph.com/cody?utm_source=server' : '/cody'
    return (
        <section className={styles.upsell}>
            <section className={styles.upsellMeta}>
                <CodyLogo withColor={true} className={styles.upsellLogo} />
                <Text className={styles.upsellTitle}>Introducing Cody: your new AI coding assistant.</Text>
                <Text className={styles.upsellDescription}>
                    Cody autocompletes single lines, or entire code blocks, in any programming language, keeping all of
                    your company’s codebase in mind.
                </Text>
                <Link to={exploreCodyLink}>Explore Cody</Link>
            </section>
            <MultiLineCompletion isLightTheme={isLightTheme} className={styles.upsellImage} />
        </section>
    )
}
