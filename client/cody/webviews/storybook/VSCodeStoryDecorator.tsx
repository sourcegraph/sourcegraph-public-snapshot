import { DecoratorFn } from '@storybook/react'

import styles from './VSCodeStoryDecorator.module.css'

/**
 * A decorator for storybooks that makes them look like they're running in VS Code.
 */
export const VSCodeStoryDecorator: DecoratorFn = story => <div className={styles.container}>{story()}</div>
