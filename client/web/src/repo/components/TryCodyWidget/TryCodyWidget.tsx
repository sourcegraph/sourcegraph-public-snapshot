import React, { useCallback, useEffect } from 'react'

import { mdiClose } from '@mdi/js'
import classNames from 'classnames'

import { useTemporarySetting } from '@sourcegraph/shared/src/settings/temporary'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { useIsLightTheme } from '@sourcegraph/shared/src/theme'
import { Button, H4, Icon, Text } from '@sourcegraph/wildcard'

import { MarketingBlock } from '../../../components/MarketingBlock'
import { EventName } from '../../../util/constants'

import styles from './TryCodyWidget.module.scss'

const GlowingCodySVG: React.FC = () => (
    <svg width="129" height="120" viewBox="0 0 129 120" fill="none" xmlns="http://www.w3.org/2000/svg">
        <g filter="url(#filter0_dd_35_36617)">
            <rect x="25.2822" y="22.9186" width="78.4355" height="74.1628" rx="6" fill="#E8D1FF" />
            <path
                fillRule="evenodd"
                clipRule="evenodd"
                d="M75.8031 39.0714C78.1652 39.0714 80.0801 40.9864 80.0801 43.3487L80.0801 53.1256C80.0801 55.488 78.1652 57.403 75.8031 57.403C73.441 57.403 71.5262 55.488 71.5262 53.1256L71.5262 43.3487C71.5262 40.9864 73.441 39.0713 75.8031 39.0714Z"
                fill="#A305E1"
            />
            <path
                fillRule="evenodd"
                clipRule="evenodd"
                d="M44.0312 50.0703C44.0312 47.708 45.9461 45.7929 48.3082 45.7929H58.0841C60.4462 45.7929 62.3611 47.708 62.3611 50.0703C62.3611 52.4327 60.4462 54.3477 58.0841 54.3477H48.3082C45.9461 54.3477 44.0312 52.4327 44.0312 50.0703Z"
                fill="#A112FF"
            />
            <path
                fillRule="evenodd"
                clipRule="evenodd"
                d="M48.9633 65.5066C47.5415 63.637 44.8751 63.263 42.9933 64.6746C41.1036 66.092 40.7206 68.7731 42.1379 70.6629L45.5594 68.0965C42.1379 70.6629 42.1391 70.6645 42.1403 70.6662L42.1429 70.6696L42.1485 70.6771L42.1616 70.6944C42.1712 70.7069 42.1825 70.7216 42.1954 70.7383C42.2212 70.7717 42.2537 70.8132 42.2929 70.8622C42.3712 70.9602 42.4763 71.0886 42.6082 71.243C42.8717 71.5514 43.2438 71.9653 43.7243 72.4487C44.6829 73.4132 46.0882 74.6697 47.9407 75.9197C51.659 78.4287 57.2127 80.9286 64.5002 80.9286C71.7878 80.9286 77.3414 78.4287 81.0598 75.9197C82.9123 74.6697 84.3175 73.4132 85.2762 72.4487C85.7567 71.9653 86.1288 71.5514 86.3923 71.243C86.5242 71.0886 86.6293 70.9602 86.7076 70.8622C86.7468 70.8132 86.7793 70.7717 86.8051 70.7383C86.818 70.7216 86.8293 70.7069 86.8388 70.6944L86.852 70.6771L86.8576 70.6696L86.8602 70.6662C86.8614 70.6645 86.8626 70.6629 83.4411 68.0965L86.8626 70.6629C88.2799 68.7731 87.8969 66.092 86.0072 64.6746C84.1254 63.2631 81.459 63.637 80.0372 65.5066C80.0348 65.5097 80.0305 65.5152 80.0243 65.5229C80.0025 65.5502 79.9575 65.6056 79.8895 65.6852C79.7533 65.8446 79.5263 66.0991 79.2097 66.4176C78.574 67.0572 77.5926 67.9394 76.2756 68.8281C73.6548 70.5965 69.7381 72.3739 64.5002 72.3739C59.2624 72.3739 55.3457 70.5965 52.7249 68.8281C51.4079 67.9394 50.4265 67.0572 49.7908 66.4176C49.4742 66.0991 49.2472 65.8446 49.111 65.6852C49.043 65.6056 48.998 65.5502 48.9762 65.5229C48.97 65.5152 48.9657 65.5097 48.9633 65.5066ZM80.0245 65.5234L80.0239 65.5242L80.0218 65.527C80.0207 65.5285 80.0195 65.53 83.3308 68.0138L80.0195 65.5301C80.0211 65.5278 80.0228 65.5256 80.0245 65.5234Z"
                fill="#A305E1"
            />
        </g>
        <defs>
            <filter
                id="filter0_dd_35_36617"
                x="11.2822"
                y="12.9186"
                width="106.436"
                height="102.163"
                filterUnits="userSpaceOnUse"
                colorInterpolationFilters="sRGB"
            >
                <feFlood floodOpacity="0" result="BackgroundImageFix" />
                <feColorMatrix
                    in="SourceAlpha"
                    type="matrix"
                    values="0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 127 0"
                    result="hardAlpha"
                />
                <feOffset dy="4" />
                <feGaussianBlur stdDeviation="3" />
                <feColorMatrix type="matrix" values="0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0.05 0" />
                <feBlend mode="normal" in2="BackgroundImageFix" result="effect1_dropShadow_35_36617" />
                <feColorMatrix
                    in="SourceAlpha"
                    type="matrix"
                    values="0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 127 0"
                    result="hardAlpha"
                />
                <feMorphology radius="3" operator="erode" in="SourceAlpha" result="effect2_dropShadow_35_36617" />
                <feOffset dy="4" />
                <feGaussianBlur stdDeviation="8.5" />
                <feColorMatrix type="matrix" values="0 0 0 0 0.556863 0 0 0 0 0.207843 0 0 0 0 0.956863 0 0 0 0.42 0" />
                <feBlend mode="normal" in2="effect1_dropShadow_35_36617" result="effect2_dropShadow_35_36617" />
                <feBlend mode="normal" in="SourceGraphic" in2="effect2_dropShadow_35_36617" result="shape" />
            </filter>
        </defs>
    </svg>
)

const AUTO_DISMISS_ON_EVENTS = new Set([
    EventName.CODY_SIDEBAR_CHAT_OPENED,
    EventName.CODY_SIDEBAR_EDIT,
    EventName.CODY_SIDEBAR_RECIPE,
    EventName.CODY_SIDEBAR_RECIPE_EXECUTED,
    EventName.CODY_SIDEBAR_SUBMIT,
])

function useTryCodyWidget(telemetryService: TelemetryProps['telemetryService']): {
    isDismissed: boolean | undefined
    onDismiss: () => void
} {
    // `isDismissed = true` maintain the initial concealment of the CTA when loading the settings
    const [isDismissed = true, setIsDismissed] = useTemporarySetting('cody.blobPageCta.dismissed', false)

    const onDismiss = useCallback(() => {
        setIsDismissed(true)
    }, [setIsDismissed])

    // Listen for telemetry events to auto dismiss the widget
    useEffect(() => {
        if (isDismissed) {
            return
        }

        return telemetryService.addEventLogListener?.(eventName => {
            if (AUTO_DISMISS_ON_EVENTS.has(eventName as EventName)) {
                onDismiss()
            }
        })
    }, [telemetryService, isDismissed, onDismiss])

    return { isDismissed, onDismiss }
}

interface TryCodyWidgetProps extends TelemetryProps {
    className?: string
    type: 'blob' | 'repo'
}

export const TryCodyWidget: React.FC<TryCodyWidgetProps> = ({ className, telemetryService, type }) => {
    const isLightTheme = useIsLightTheme()
    const { isDismissed, onDismiss } = useTryCodyWidget(telemetryService)
    useEffect(() => {
        if (isDismissed) {
            return
        }
        const eventPage = type === 'blob' ? 'BlobPage' : 'RepoPage'
        telemetryService.log(EventName.TRY_CODY_WEB_ONBOARDING_DISPLAYED, { type: eventPage }, { type: eventPage })
    }, [isDismissed, telemetryService, type])

    if (isDismissed) {
        return null
    }

    const { title, useCases, image } =
        type === 'blob'
            ? {
                  title: 'Try Cody on public code',
                  useCases: ['Select code in the file below', 'Select an action with Cody widget'],
                  image: `https://storage.googleapis.com/sourcegraph-assets/app-images/cody-action-bar-${
                      isLightTheme ? 'light' : 'dark'
                  }.png`,
              }
            : {
                  title: 'Try Cody AI assist on this repo',
                  useCases: [
                      'Click the Ask Cody button above and to the right of this banner',
                      'Ask Cody a question like “Explain the structure of this repository”',
                  ],
                  image: `https://storage.googleapis.com/sourcegraph-assets/app-images/cody-chat-banner-image-${
                      isLightTheme ? 'light' : 'dark'
                  }.png`,
              }

    return (
        <MarketingBlock
            wrapperClassName={classNames(className, type === 'blob' ? styles.blobCardWrapper : styles.repoCardWrapper)}
            contentClassName={classNames(
                'd-flex position-relative pb-0 overflow-auto justify-content-between',
                styles.card
            )}
            variant="thin"
        >
            <div className="d-flex pb-3">
                <div>
                    <GlowingCodySVG />
                </div>
                <div className="d-flex flex-column flex-grow-1 justify-content-center">
                    <H4 className={styles.cardTitle}>{title}</H4>
                    <ol className={classNames('m-0 pl-4 fs-6', styles.cardList)}>
                        {useCases.map(useCase => (
                            <Text key={useCase} as="li">
                                {useCase}
                            </Text>
                        ))}
                    </ol>
                </div>
            </div>
            <div className={classNames('d-flex justify-content-center', styles.cardImages)}>
                <img src={image} alt="Cody" className={styles.cardImage} />
            </div>
            <Button className={classNames(styles.closeButton, 'position-absolute mt-2')} onClick={onDismiss}>
                <Icon svgPath={mdiClose} aria-label="Close try Cody widget" />
            </Button>
        </MarketingBlock>
    )
}
