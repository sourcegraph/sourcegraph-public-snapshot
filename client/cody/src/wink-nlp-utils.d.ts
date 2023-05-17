declare module 'wink-nlp-utils' {
    declare namespace winkUtils {
        namespace string {
            function tokenize0(s: string): string[]
            function stem(word: string): string | undefined
        }

        namespace tokens {
            function removeWords(words: string[]): string[]
        }
    }

    export default winkUtils
}
