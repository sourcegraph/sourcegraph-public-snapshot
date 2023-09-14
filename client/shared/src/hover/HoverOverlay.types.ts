import type { HoverMerged } from '@sourcegraph/client-api'
import type { HoverOverlayProps as GenericHoverOverlayProps } from '@sourcegraph/codeintellify'

import type { ActionItemAction } from '../actions/ActionItem'
import type { FileSpec, RepoSpec, ResolvedRevisionSpec, RevisionSpec } from '../util/url'

export type HoverContext = RepoSpec & RevisionSpec & FileSpec & ResolvedRevisionSpec

export interface HoverOverlayBaseProps extends GenericHoverOverlayProps<HoverContext, HoverMerged, ActionItemAction> {}
