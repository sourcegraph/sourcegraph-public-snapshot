import { CodeHosts, RepogroupMetadata } from './types'
import { SearchPatternType } from '../../../shared/src/graphql/schema'
import * as React from 'react'

export const android: RepogroupMetadata = {
    title: 'Android',
    name: 'android',
    url: '/android',
    repositories: [
        { name: 'kotlin/kotlinx.serialization', codehost: CodeHosts.GITHUB },
        { name: 'kotlin/dokka', codehost: CodeHosts.GITHUB },
        { name: 'kotlin/kotlinx.coroutines', codehost: CodeHosts.GITHUB },
        { name: 'kotlin/kotlin-fullstack-sample', codehost: CodeHosts.GITHUB },
        { name: 'google/dagger', codehost: CodeHosts.GITHUB },
        { name: 'square/retrofit', codehost: CodeHosts.GITHUB },
        { name: 'square/okhttp', codehost: CodeHosts.GITHUB },
        { name: 'square/moshi', codehost: CodeHosts.GITHUB },
        { name: 'jgilfelt/chuck', codehost: CodeHosts.GITHUB },
        { name: 'google/gson', codehost: CodeHosts.GITHUB },
        { name: 'bumptech/glide', codehost: CodeHosts.GITHUB },
        { name: 'ReactiveX/RxJava', codehost: CodeHosts.GITHUB },
        { name: 'JakeWharton/timber', codehost: CodeHosts.GITHUB },
        { name: 'JakeWharton/ThreeTenABP', codehost: CodeHosts.GITHUB },
        { name: 'cashapp/sqldelight', codehost: CodeHosts.GITHUB },
        { name: 'JakeWharton/SdkSearch', codehost: CodeHosts.GITHUB },
        { name: 'JakeWharton/butterknife', codehost: CodeHosts.GITHUB },
        { name: 'JakeWharton/Telecine', codehost: CodeHosts.GITHUB },
        { name: 'YiiGuxing/TranslationPlugin', codehost: CodeHosts.GITHUB },
        { name: 'pbreault/adb-idea', codehost: CodeHosts.GITHUB },
        { name: 'JetBrains/kotlin', codehost: CodeHosts.GITHUB },
        { name: 'JetBrains/intellij-plugins', codehost: CodeHosts.GITHUB },
        { name: 'JetBrains/intellij-community', codehost: CodeHosts.GITHUB },
        { name: 'JetBrains/kotlin-native', codehost: CodeHosts.GITHUB },
        { name: 'JetBrains/kotlin-web-site', codehost: CodeHosts.GITHUB },
        { name: 'JetBrains/ideavim', codehost: CodeHosts.GITHUB },
        { name: 'android/kotlin', codehost: CodeHosts.GITHUB },
        { name: 'android/user-interface-samples', codehost: CodeHosts.GITHUB },
        { name: 'android/views-widgets-samples', codehost: CodeHosts.GITHUB },
        { name: 'android/animation-samples', codehost: CodeHosts.GITHUB },
        { name: 'android/play-billing-samples', codehost: CodeHosts.GITHUB },
        { name: 'android/architecture-samples', codehost: CodeHosts.GITHUB },
        { name: 'android/location-samples', codehost: CodeHosts.GITHUB },
        { name: 'android/camera-samples', codehost: CodeHosts.GITHUB },
        { name: 'android/compose-samples', codehost: CodeHosts.GITHUB },
        { name: 'android/ndk-samples', codehost: CodeHosts.GITHUB },
        { name: 'android/plaid', codehost: CodeHosts.GITHUB },
        { name: 'android/topeka', codehost: CodeHosts.GITHUB },
        { name: 'android/testing-samples', codehost: CodeHosts.GITHUB },
        { name: 'android/databinding-samples', codehost: CodeHosts.GITHUB },
        { name: 'android/kotlin-guides', codehost: CodeHosts.GITHUB },
        { name: 'android/android-test', codehost: CodeHosts.GITHUB },
        { name: 'android/uamp', codehost: CodeHosts.GITHUB },
        { name: 'android/sunflower', codehost: CodeHosts.GITHUB },
        { name: 'skydoves/TransformationLayout', codehost: CodeHosts.GITHUB },
        { name: 'codepath/android_guides', codehost: CodeHosts.GITHUB },
        { name: 'google/iosched', codehost: CodeHosts.GITHUB },
        { name: 'futurice/android-best-practices', codehost: CodeHosts.GITHUB },
        { name: 'lgvalle/Material-Animations', codehost: CodeHosts.GITHUB },
        { name: 'wasabeef/awesome-android-ui', codehost: CodeHosts.GITHUB },
        { name: 'nisrulz/android-tips-tricks', codehost: CodeHosts.GITHUB },
        { name: 'MindorksOpenSource/from-java-to-kotlin', codehost: CodeHosts.GITHUB },
        { name: 'dbacinski/Design-Patterns-In-Kotlin', codehost: CodeHosts.GITHUB },
        { name: 'Solido/awesome-flutter', codehost: CodeHosts.GITHUB },
        { name: 'iampawan/FlutterExampleApps', codehost: CodeHosts.GITHUB },
        { name: 'mitesh77/Best-Flutter-UI-Templates', codehost: CodeHosts.GITHUB },
    ],
    description: 'Interesting search examples in popular Android repositories.',
    examples: [
        {
            title: 'Find usage examples of the OnClickListener function in Kotlin’s syntax:',
            exampleQuery: (
                <>
                    <span className="repogroup-page__keyword-text">file:</span>
                    {'.kt .setOnClickListener {:[function]}'}
                </>
            ),
            rawQuery: 'file:.kt .setOnClickListener {:[function]}',
            patternType: SearchPatternType.structural,
        },
        {
            title: 'Find intent filter examples in Android Manifest XML files:',
            exampleQuery: (
                <>
                    <span className="repogroup-page__keyword-text">lang:</span>
                    {'.xml <intent-filter> :[string] </intent-filter>'}
                </>
            ),
            rawQuery: 'lang:.xml <intent-filter> :[string] </intent-filter>',
            patternType: SearchPatternType.structural,
        },
        {
            title: 'Detect nested LinearLayouts in your XML layout files:',
            exampleQuery: (
                <>
                    <span className="repogroup-page__keyword-text">file:</span>
                    {'xml <LinearLayout :[0] <LinearLayout :[1] </LinearLayout> :[2] </LinearLayout>'}
                </>
            ),
            rawQuery: 'file:xml <LinearLayout :[0] <LinearLayout :[1] </LinearLayout> :[2] </LinearLayout>',
            patternType: SearchPatternType.structural,
        },
        {
            title: 'Find try-catch blocks to see how errors are caught:',
            exampleQuery: <>{'try {:[0]} catch (:[1]) {:[2]} finally {:[3]}'}</>,
            rawQuery: 'try {:[0]} catch (:[1]) {:[2]} finally {:[3]}',
            patternType: SearchPatternType.structural,
        },
        //         {
        //             title: 'Switch statements in Java',
        //             exampleQuery:
        //                 'http.Transport{:[_], MaxIdleConns: :[idleconns], :[_]} <span className="repogroup-page__keyword-text">-file:</span>vendor <span class="repogroup-page__keyword-text">lang:</span>go',
        //         },
    ],
    homepageDescription: 'Interesting examples of Go.',
    homepageIcon: 'https://code.benco.io/icon-collection/logos/go-lang.svg',
}
