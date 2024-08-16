import "@typespec/http";
import "@typespec/rest";

using TypeSpec.Http;
using TypeSpec.Rest;

@service({
    title: "Cody Service",
})
namespace CodyService;

@route("/models")
interface Models {
    @get
    @summary("Lists the currently available models, and provides basic information about each one such as the owner and availability.")
    list(): ListModelsResponse | Error;

    @get
    @summary("Retrieves a model instance, providing basic information about the model such as the owner and permissioning.")
    retrieveModel(@path modelId: string): Model | Error;
}

@doc("Describes an OpenAI model offering that can be used with the API.")
model Model {
    @doc("The model identifier, which can be referenced in the API endpoints.")
    id: string;

    @doc("The object type, which is always \"model\".")
    object: "model";

    @doc("The Unix timestamp (in seconds) when the model was created.")
    created: int64;

    @doc("The organization that owns the model.")
    owned_by: string;
}

model ListModelsResponse {
    object: "list";
    data: Model[];
}

@error
model Error {
    code: int32;
    message: string;
}
