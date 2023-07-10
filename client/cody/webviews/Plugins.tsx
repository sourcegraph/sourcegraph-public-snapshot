import { defaultPlugins } from '../src/plugins/examples'

import styles from './Plugins.module.css'

export const Plugins: React.FunctionComponent = () => (
    <div className={styles.container}>
        <input placeholder="Search plugins @installed, @enabled..." />
        <div className={styles.divider} />
        {defaultPlugins.map(plugin => (
            <section className={styles.plugin} key={plugin.name}>
                <div className={styles.pluginHeader}>
                    <input type="checkbox" checked={true} />
                    <h3>{plugin.name}</h3>
                </div>
                <p className={styles.pluginDescription}>{plugin.description}</p>
            </section>
        ))}
    </div>
)
