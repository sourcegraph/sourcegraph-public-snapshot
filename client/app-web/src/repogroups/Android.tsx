import { RepogroupMetadata } from './types'
import { SearchPatternType } from '../graphql-operations'

export const android: RepogroupMetadata = {
    title: 'Android',
    name: 'android',
    url: '/android',
    description: 'Use these search suggestions to explore popular Android repositories on GitHub.',
    examples: [
        {
            title: 'Find intent filter examples in Android Manifest XML files',
            query: 'lang:xml <intent-filter> :[string] </intent-filter>',
            description:
                'Intent filters specify the type of intents a component would like to receive. An intent filter can accept three types of elements - <action>, <category> and <data> elements.',
            patternType: SearchPatternType.structural,
        },
        {
            title: 'Find try-catch blocks to see how errors are caught',
            query: 'try {:[0]} catch (:[1]) {:[2]} finally {:[3]}',
            patternType: SearchPatternType.structural,
        },
        {
            title: 'Examine and optimize your layout by detecting nested LinearLayouts',
            description: `LinearLayout can lead to an excessively deep view hierarchy. Nesting several instances of
            LinearLayout that use the layout_weight parameter can be especially expensive as each child needs to be measured twice. This is particularly
            important when the layout is inflated repeatedly, such as when used in a ListView or GridView.`,
            query: '<LinearLayout:[_]>:[_]<LinearLayout:[_]>:[_]</LinearLayout>:[_]</LinearLayout>',
            patternType: SearchPatternType.structural,
        },
        {
            title: 'Find usage examples of the OnClickListener function in Kotlin’s syntax',
            query: 'file:.kt .setOnClickListener {:[function]}',
            patternType: SearchPatternType.structural,
        },
    ],
    homepageDescription: 'Explore popular Android repositories.',
    homepageIcon: 'https://code.benco.io/icon-collection/logos/android-1.svg',
}
