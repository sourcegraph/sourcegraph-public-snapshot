package completions

import (
	"strings"
	"testing"

	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"
)

func TestGoogleRequestGetTokenCount(t *testing.T) {
	logger := logtest.Scoped(t)

	t.Run("streaming", func(t *testing.T) {
		req := googleRequest{stream: true}
		r := strings.NewReader(googleStreamingResponse)
		handler := &GoogleHandlerMethods{}
		promptUsage, completionUsage := handler.parseResponseAndUsage(logger, req, r)

		assert.Equal(t, 21, promptUsage.tokens)
		assert.Equal(t, 87, completionUsage.tokens)
	})

	t.Run("non-streaming", func(t *testing.T) {
		req := googleRequest{stream: false}
		r := strings.NewReader(googleNonStreamingResponse)
		handler := &GoogleHandlerMethods{}
		promptUsage, completionUsage := handler.parseResponseAndUsage(logger, req, r)

		assert.Equal(t, 16, promptUsage.tokens)
		assert.Equal(t, 19, completionUsage.tokens)
	})
}

var googleStreamingResponse = `data: {"candidates": [{"content": {"parts": [{"text": "def"}],"role": "model"},"finishReason": "STOP","index": 0}],"usageMetadata": {"promptTokenCount": 21,"candidatesTokenCount": 1,"totalTokenCount": 22}}

data: {"candidates": [{"content": {"parts": [{"text": " bubble_sort(list1):\n  n = len(list1)"}],"role": "model"},"finishReason": "STOP","index": 0,"safetyRatings": [{"category": "HARM_CATEGORY_SEXUALLY_EXPLICIT","probability": "NEGLIGIBLE"},{"category": "HARM_CATEGORY_HATE_SPEECH","probability": "NEGLIGIBLE"},{"category": "HARM_CATEGORY_HARASSMENT","probability": "NEGLIGIBLE"},{"category": "HARM_CATEGORY_DANGEROUS_CONTENT","probability": "NEGLIGIBLE"}]}],"usageMetadata": {"promptTokenCount": 21,"candidatesTokenCount": 17,"totalTokenCount": 38}}

data: {"candidates": [{"content": {"parts": [{"text": "\n  for i in range(n-1):\n    for j in"}],"role": "model"},"finishReason": "STOP","index": 0,"safetyRatings": [{"category": "HARM_CATEGORY_SEXUALLY_EXPLICIT","probability": "NEGLIGIBLE"},{"category": "HARM_CATEGORY_HATE_SPEECH","probability": "NEGLIGIBLE"},{"category": "HARM_CATEGORY_HARASSMENT","probability": "NEGLIGIBLE"},{"category": "HARM_CATEGORY_DANGEROUS_CONTENT","probability": "NEGLIGIBLE"}]}],"usageMetadata": {"promptTokenCount": 21,"candidatesTokenCount": 31,"totalTokenCount": 52}}

data: {"candidates": [{"content": {"parts": [{"text": " range(n-i-1):\n      if list1[j] \u003e list1[j+1]:\n        list1[j], list"}],"role": "model"},"finishReason": "STOP","index": 0,"safetyRatings": [{"category": "HARM_CATEGORY_SEXUALLY_EXPLICIT","probability": "NEGLIGIBLE"},{"category": "HARM_CATEGORY_HATE_SPEECH","probability": "NEGLIGIBLE"},{"category": "HARM_CATEGORY_HARASSMENT","probability": "NEGLIGIBLE"},{"category": "HARM_CATEGORY_DANGEROUS_CONTENT","probability": "NEGLIGIBLE"}]}],"usageMetadata": {"promptTokenCount": 21,"candidatesTokenCount": 63,"totalTokenCount": 84}}

data: {"candidates": [{"content": {"parts": [{"text": "1[j+1] = list1[j+1], list1[j]\n  return list1\n"}],"role": "model"},"finishReason": "STOP","index": 0,"safetyRatings": [{"category": "HARM_CATEGORY_SEXUALLY_EXPLICIT","probability": "NEGLIGIBLE"},{"category": "HARM_CATEGORY_HATE_SPEECH","probability": "NEGLIGIBLE"},{"category": "HARM_CATEGORY_HARASSMENT","probability": "NEGLIGIBLE"},{"category": "HARM_CATEGORY_DANGEROUS_CONTENT","probability": "NEGLIGIBLE"}],"citationMetadata": {"citationSources": [{"startIndex": 1,"endIndex": 185,"uri": "https://github.com/Feng080412/Searches-and-sorts","license": ""}]}}],"usageMetadata": {"promptTokenCount": 21,"candidatesTokenCount": 87,"totalTokenCount": 108}}`

var googleNonStreamingResponse = `{
  "candidates": [
    {
      "content": {
        "parts": [
          {
            "text": "My name is Cody. Nice to meet you! What's your name? \n"
          }
        ],
        "role": "model"
      },
      "finishReason": "STOP",
      "index": 0,
      "safetyRatings": [
        {
          "category": "HARM_CATEGORY_SEXUALLY_EXPLICIT",
          "probability": "NEGLIGIBLE"
        },
        {
          "category": "HARM_CATEGORY_HATE_SPEECH",
          "probability": "NEGLIGIBLE"
        },
        {
          "category": "HARM_CATEGORY_HARASSMENT",
          "probability": "NEGLIGIBLE"
        },
        {
          "category": "HARM_CATEGORY_DANGEROUS_CONTENT",
          "probability": "NEGLIGIBLE"
        }
      ]
    }
  ],
  "usageMetadata": {
    "promptTokenCount": 16,
    "candidatesTokenCount": 19,
    "totalTokenCount": 35
  }
}`
