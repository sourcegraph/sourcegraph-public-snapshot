POST https://localhost:3443/api/openai/v1/chat/completions
Content-Type: application/json
Authorization: Bearer {{token}}
{
  "model": "anthropic::unknown::claude-3-sonnet-20240229",
  "messages": [
    {
      "role": "user",
      "content": [
        {
          "type": "text",
          "text": "Respond with \"no\" and nothing else"
        }
      ]
    }
  ],
  "temperature": 1,
  "max_tokens": 256,
  "top_p": 1,
  "frequency_penalty": 0,
  "presence_penalty": 0
}
