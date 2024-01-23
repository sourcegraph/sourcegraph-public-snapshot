import type { FC } from 'react'

import { H1, H3, Text, Link } from '@sourcegraph/wildcard'

import { CodyLogo } from '../components/CodyLogo'

import styles from './CodyUpsellPage.module.scss'

interface CodyUpsellPageProps { }

export const CodyUpsellPage: FC<CodyUpsellPageProps> = () => (
  <div className={styles.container}>
    <HeroBanner />

    <div className={styles.noAccessBanner}>
      <Text className={styles.noAccessBannerHeader}>
        You currently don’t have access to Cody for Enterprise! 😞
      </Text>
      <small className={styles.noAccessBannerDescription}>
        If you’d to try it out, ask your admin for access.
      </small>
    </div>

    <div className={styles.codyIntro}>
      <div className={styles.codyIntroTextContainer}>
        <CodyLogo withColor={true} className={styles.codyIntroLogo} />
        <H3 className={styles.codyIntroHeader}>Introducting Cody: your new AI coding assistant.</H3>
        <Text className={styles.codyIntroDescription}>
          Cody autocompletes single lines, or entire code blocks, in any programming language, keeping all of
          your company’s codebase in mind.
        </Text>

        <Link to="/help/cody/overview" className={styles.codyIntroLink}>
          Explore Cody
        </Link>
      </div>

      <Text>Image</Text>
    </div>
  </div >
)

const HeroBanner: FC = () => (
  <div className={styles.heroContainer}>
    <div className={styles.heroDetails}>
      <CodyLogo withColor={true} className={styles.heroCodyLogo} />

      <H1 className="my-2">
        Meet your new AI assistant, <span className={styles.heroCodyHighlight}>Cody.</span>
      </H1>

      <Text className={styles.heroCodyDescription}>
        Cody is a coding AI assistant that uses AI and a deep understanding of your codebase to help you write
        and understand code faster.
      </Text>
    </div>
  </div>
)
