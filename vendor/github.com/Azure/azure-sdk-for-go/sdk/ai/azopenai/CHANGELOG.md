# Release History

## 0.5.0 (2024-03-05)

### Features Added

- Updating to the `2024-02-15-preview` API version.
- `GetAudioSpeech` enables translating text to speech.

### Breaking Changes

- Citations, previously returned as an unparsed JSON blob, are now deserialized into a real type in `ChatResponseMessage.Citations`.
- `AzureCognitiveSearchChatExtensionConfiguration` has been renamed to `AzureSearchChatExtensionConfiguration`.
- `AzureCognitiveSearchChatExtensionParameters` has been renamed to `AzureSearchChatExtensionParameters`.

## 0.4.1 (2024-01-16)

### Bugs Fixed

- `AudioTranscriptionOptions.Filename` and `AudioTranslationOptions.Filename` fields are now properly propagated, allowing 
  for disambiguating the format of an audio file when OpenAI can't detect it. (PR#22210) 

## 0.4.0 (2023-12-11)

Support for many of the features mentioned in OpenAI's November Dev Day and Microsoft's 2023 Ignite conference

### Features Added

- Chat completions has been extended to accomodate new features:
  - Parallel function calling via Tools. See the function `ExampleClient_GetChatCompletions_functions` in `example_client_getchatcompletions_extensions_test.go` for an example of specifying a Tool.
  - "JSON mode", via `ChatCompletionOptions.ResponseFormat` for guaranteed function outputs.
- ChatCompletions can now be used with both text and images using `gpt-4-vision-preview`.
  - Azure enhancements to `gpt-4-vision-preview` results that include grounding and OCR features
- GetImageGenerations now works with DallE-3.
- `-1106` model feature support for `gpt-35-turbo` and `gpt-4-turbo`, including use of a seed via `ChatCompletionsOptions.Seed` and system fingerprints returned in `ChatCompletions.SystemFingerprint`.
- `dall-e-3` image generation capabilities via `GetImageGenerations`, featuring higher model quality, automatic prompt revisions by `gpt-4`, and customizable quality/style settings

### Breaking Changes

- `azopenai.KeyCredential` has been replaced by [azcore.KeyCredential](https://pkg.go.dev/github.com/Azure/azure-sdk-for-go/sdk/azcore#KeyCredential).
- `Deployment` has been renamed to `DeploymentName` throughout all APIs.
- `CreateImage` has been replaced with `GetImageGenerations`.
- `ChatMessage` has been split into per-role types. The function `ExampleClient_GetChatCompletions` in `example_client_getcompletions_test.go` shows an example of this.

## 0.3.0 (2023-09-26)

### Features Added
- Support for Whisper audio APIs for transcription and translation using `GetAudioTranscription` and `GetAudioTranslation`.

### Breaking Changes
- ChatChoiceContentFilterResults content filtering fields are now all typed as ContentFilterResult, instead of unique types for each field.
- `PromptAnnotations` renamed to `PromptFilterResults` in `ChatCompletions` and `Completions`.

## 0.2.0 (2023-08-28)

### Features Added

- ChatCompletions supports Azure OpenAI's newest feature to use Azure OpenAI with your own data. See `example_client_getchatcompletions_extensions_test.go`
  for a working example. (PR#21426)

### Breaking Changes

- ChatCompletionsOptions, CompletionsOptions, EmbeddingsOptions `DeploymentID` field renamed to `Deployment`.
- Method `Close()` on `EventReader[T]` now returns an error.

### Bugs Fixed

- EventReader, used by GetChatCompletionsStream and GetCompletionsStream for streaming results, would not return an 
  error if the underlying Body reader was closed or EOF'd before the actual DONE: token arrived. This could result in an
  infinite loop for callers. (PR#21323)

## 0.1.1 (2023-07-26)

### Breaking Changes

- Moved from `sdk/cognitiveservices/azopenai` to `sdk/ai/azopenai`.

## 0.1.0 (2023-07-20)

* Initial release of the `azopenai` library
