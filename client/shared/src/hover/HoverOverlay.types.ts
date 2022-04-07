import { NotificationType } from 'sourcegraph'

import { HoverMerged } from '@sourcegraph/client-api'
import { HoverOverlayProps as GenericHoverOverlayProps } from '@sourcegraph/codeintellify'
import { AlertProps } from '@sourcegraph/wildcard'

import { ActionItemAction } from '../actions/ActionItem'
import { FileSpec, RepoSpec, ResolvedRevisionSpec, RevisionSpec } from '../util/url'

export type HoverContext = RepoSpec & RevisionSpec & FileSpec & ResolvedRevisionSpec

export interface HoverOverlayBaseProps extends GenericHoverOverlayProps<HoverContext, HoverMerged, ActionItemAction> {}

export type GetAlertClassName = (
    kind: Exclude<NotificationType, NotificationType.Log | NotificationType.Success>
) => string | undefined

export type GetAlertVariant = (
    kind: Exclude<NotificationType, NotificationType.Log | NotificationType.Success>
) => AlertProps['variant']
