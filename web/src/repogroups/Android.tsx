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
            title: 'Search for usages of the Retry-After header in non-vendor Go files:',
            exampleQuery: (
                <>
                    <span className="repogroup-page__keyword-text">file:</span>.go{' '}
                    <span className="repogroup-page__keyword-text">-file:</span>vendor/ Retry-After
                </>
            ),
            rawQuery: 'file:.go -file:vendor/ Retry-After',
            patternType: SearchPatternType.literal,
        },
        {
            title: 'Find examples of sending JSON in a HTTP POST request:',
            exampleQuery: (
                <>
                    repogroup:goteam <span className="repogroup-page__keyword-text">file:</span>.go http.Post json'
                </>
            ),
            rawQuery: 'repogroup:goteam file:.go http.Post json',
            patternType: SearchPatternType.literal,
        },
        {
            title: 'Find error handling examples in Go',
            exampleQuery: (
                <>
                    {'if err != nil {:[_]}'} <span className="repogroup-page__keyword-text">lang:</span>go'
                </>
            ),
            rawQuery: 'if err != nil {:[_]} lang:go',
            patternType: SearchPatternType.structural,
        },
        {
            title: 'Find usage examples of cmp.Diff with options',
            exampleQuery: (
                <>
                    <span className="repogroup-page__keyword-text">lang:go</span> cmp.Diff(:[_], :[_], :[opts])'
                </>
            ),
            rawQuery: 'lang:go cmp.Diff(:[_], :[_], :[opts])',
            patternType: SearchPatternType.structural,
        },
        {
            title: 'Find examples for setting timeouts on http.Transport',
            exampleQuery: (
                <>
                    {'http.Transport{:[_], MaxIdleConns: :[idleconns], :[_]}'}{' '}
                    <span className="repogroup-page__keyword-text">-file:</span>vendor{' '}
                    <span className="repogroup-page__keyword-text">lang:</span>go'
                </>
            ),
            rawQuery: 'http.Transport{:[_], MaxIdleConns: :[idleconns], :[_]} -file:vendor lang:go',
            patternType: SearchPatternType.structural,
        },
        {
            title: 'Find examples of Switch statements in Go',
            exampleQuery: (
                <>
                    {'switch :[_] := :[_].(type) { :[string] }'}
                    <span className="repogroup-page__keyword-text">lang:</span>go{' '}
                    <span className="repogroup-page__keyword-text">count:</span>1000'
                </>
            ),
            rawQuery: 'switch :[_] := :[_].(type) { :[string] } lang:go count:1000',
            patternType: SearchPatternType.structural,
        },
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
