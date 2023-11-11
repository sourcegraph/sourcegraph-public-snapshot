import type { Meta } from '@storybook/react'

import { BrandedStory } from '../../../../../stories/BrandedStory'

import { ScrollBox } from './ScrollBox'

const meta: Meta = {
    title: 'wildcard/Charts/Core',
    decorators: [story => <BrandedStory>{() => <div className="container mt-3">{story()}</div>}</BrandedStory>],
}

export default meta

export const ScrollBoxDemo = () => (
    <ScrollBox style={{ height: 400, width: 200, border: '1px solid var(--border-color)' }}>
        Sorokin's works, bright and striking examples of underground culture, were banned during the Soviet period. His
        first publication in the USSR appeared in November 1989, when the Riga-based Latvian magazine Rodnik (Spring)
        presented a group of Sorokin's stories. Soon after, his stories appeared in Russian literary miscellanies and
        magazines Tretya Modernizatsiya (The Third Modernization), Mitin Zhurnal (Mitya's Journal), Konets Veka (End of
        the Century), and Vestnik Novoy Literatury (Bulletin of the New Literature). In 1992, Russian publishing house
        Russlit published Sbornik Rasskazov (Collected Stories) – Sorokin's first book to be nominated for a Russian
        Booker Prize.[4] In September 2001, Vladimir Sorokin received the People's Booker Prize; two months later, he
        was presented with the Andrei Bely Prize for outstanding contributions to Russian literature. In 2002, there was
        a protest against his book Blue Lard, and he was investigated for pornography. Sorokin's books have been
        translated into English, Portuguese, Spanish, French, German, Dutch, Finnish, Swedish, Norwegian, Danish,
        Italian, Polish, Japanese, Serbian, Korean, Romanian, Estonian, Slovak, Czech, Hungarian, Croatian and
        Slovenian, and are available through a number of prominent publishing houses, including Gallimard, Fischer,
        DuMont, BV Berlin, Haffman, Mlinarec & Plavic and Verlag der Autoren.
    </ScrollBox>
)
