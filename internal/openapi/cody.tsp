import "@typespec/http";
import "@typespec/rest";

using TypeSpec.Http;
using TypeSpec.Rest;
@route("/.api/cody")
interface CodyService {
    @route("search")
    @get
    search(
        @query repos: string[],
        @query query: string,
        @query codeResultsCount: int32,
        @query textResultsCount: int32,
    ): CodySearchResponse;
}

model CodySearchResponse {
    results: FileChunkContext[];
}

model FileChunkContext {
    blob: BlobInfo;
    startLine: int32;
    endLine: int32;
    chunkContent: string;
}

model BlobInfo {
    path: string;
    repository: RepositoryInfo;
    commit: CommitInfo;
    url: string;
}

model RepositoryInfo {
    id: string;
    name: string;
}

model CommitInfo {
    oid: string;
}
