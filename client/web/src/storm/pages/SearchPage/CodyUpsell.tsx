import type { FC } from 'react'

import { useIsLightTheme } from '@sourcegraph/shared/src/theme'

import { MultiLineCompletion } from './MultilineCompletion'

import styles from './CodyUpsell.module.scss'

export const CodyUpsell: FC = () => {
    const isLightTheme = useIsLightTheme()
    return (
        <section className={styles.upsell}>
            <section>Cody Upsell</section>
            <MultiLineCompletion isLightTheme={isLightTheme} className="" />
        </section>
    )
}
