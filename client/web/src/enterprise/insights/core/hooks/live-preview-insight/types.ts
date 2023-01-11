export enum LivePreviewStatus {
    Intact,
    Loading,
    Error,
    Data,
}

export type State<D> =
    | { status: LivePreviewStatus.Data; data: D }
    | { status: LivePreviewStatus.Error; error: Error }
    | { status: LivePreviewStatus.Loading }
    | { status: LivePreviewStatus.Intact }
