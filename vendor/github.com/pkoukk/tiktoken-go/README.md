# tiktoken-go
[简体中文](./README_zh-hans.md)

OpenAI's tiktoken in Go. 

Tiktoken is a fast BPE tokeniser for use with OpenAI's models.

This is a port of the original [tiktoken](https://github.com/openai/tiktoken).  

# Usage
## Install

```bash
go get github.com/pkoukk/tiktoken-go
```
## Cache
Tiktoken-go has the same cache mechanism as the original Tiktoken library.  

You can set the cache directory by using the environment variable TIKTOKEN_CACHE_DIR.   

Once this variable is set, tiktoken-go will use this directory to cache the token dictionary.   

If you don't set this environment variable, tiktoken-go will download the dictionary each time you initialize an encoding for the first time.  

## Alternative BPE loaders
If you don't want to use cache or download the dictionary each time, you can use alternative BPE loader.

Just call `tiktoken.SetBpeLoader` before calling `tiktoken.GetEncoding` or `tiktoken.EncodingForModel`.

`BpeLoader` is an interface, you can implement your own BPE loader by implementing this interface.

### Offline BPE loader
The offline BPE loader loads the BPE dictionary from embed files, it helps if you don't want to download the dictionary at runtime.  

Due to the size of the BPE dictionary, this loader is in other project.

Include if you require this loader: [tiktoken_loader](https://github.com/pkoukk/tiktoken-go-loader)

## Examples
### Get Token By Encoding

```go
package main

import (
    "fmt"
    "github.com/pkoukk/tiktoken-go"
)

func main()  {
	text := "Hello, world!"
	encoding := "cl100k_base"

	// if you don't want download dictionary at runtime, you can use offline loader
	// tiktoken.SetBpeLoader(tiktoken_loader.NewOfflineLoader())
	tke, err := tiktoken.GetEncoding(encoding)
	if err != nil {
		err = fmt.Errorf("getEncoding: %v", err)
		return
	}

	// encode
	token := tke.Encode(text, nil, nil)

	//tokens
	fmt.Println((token))
	// num_tokens
	fmt.Println(len(token))
}
```

### Get Token By Model

```go
package main

import (
    "fmt"
    "github.com/pkoukk/tiktoken-go"
)

func main()  {
	text := "Hello, world!"
	encoding := "gpt-3.5-turbo"

	tkm, err := tiktoken.EncodingForModel(encoding)
	if err != nil {
		err = fmt.Errorf("getEncoding: %v", err)
		return
	}

	// encode
	token := tkm.Encode(text, nil, nil)

	// tokens
	fmt.Println(token)
	// num_tokens
	fmt.Println(len(token))
}
```

### Counting Tokens For Chat API Calls
Below is an example function for counting tokens for messages passed to gpt-3.5-turbo or gpt-4.

The following code was written based on [openai-cookbook](https://github.com/openai/openai-cookbook/blob/main/examples/How_to_count_tokens_with_tiktoken.ipynb)  examples at `Wednesday, 28 June 2023`.

Please note that the token calculation method for the message may change at any time, so this code may not necessarily be applicable in the future.

If you need accurate calculation, please refer to the official documentation.

If you find that this code is no longer applicable, please feel free to submit a PR or Issue.


```go
package main

import (
	"fmt"

	"github.com/pkoukk/tiktoken-go"
	"github.com/sashabaranov/go-openai"
)

// OpenAI Cookbook: https://github.com/openai/openai-cookbook/blob/main/examples/How_to_count_tokens_with_tiktoken.ipynb
func NumTokensFromMessages(messages []openai.ChatCompletionMessage, model string) (numTokens int) {
	tkm, err := tiktoken.EncodingForModel(model)
	if err != nil {
		err = fmt.Errorf("encoding for model: %v", err)
		log.Println(err)
		return
	}

	var tokensPerMessage, tokensPerName int
	switch model {
	case "gpt-3.5-turbo-0613",
		"gpt-3.5-turbo-16k-0613",
		"gpt-4-0314",
		"gpt-4-32k-0314",
		"gpt-4-0613",
		"gpt-4-32k-0613":
		tokensPerMessage = 3
		tokensPerName = 1
	case "gpt-3.5-turbo-0301":
		tokensPerMessage = 4 // every message follows <|start|>{role/name}\n{content}<|end|>\n
		tokensPerName = -1   // if there's a name, the role is omitted
	default:
		if strings.Contains(model, "gpt-3.5-turbo") {
			log.Println("warning: gpt-3.5-turbo may update over time. Returning num tokens assuming gpt-3.5-turbo-0613.")
			return NumTokensFromMessages(messages, "gpt-3.5-turbo-0613")
		} else if strings.Contains(model, "gpt-4") {
			log.Println("warning: gpt-4 may update over time. Returning num tokens assuming gpt-4-0613.")
			return NumTokensFromMessages(messages, "gpt-4-0613")
		} else {
			err = fmt.Errorf("num_tokens_from_messages() is not implemented for model %s. See https://github.com/openai/openai-python/blob/main/chatml.md for information on how messages are converted to tokens.", model)
			log.Println(err)
			return
		}
	}

	for _, message := range messages {
		numTokens += tokensPerMessage
		numTokens += len(tkm.Encode(message.Content, nil, nil))
		numTokens += len(tkm.Encode(message.Role, nil, nil))
		numTokens += len(tkm.Encode(message.Name, nil, nil))
		if message.Name != "" {
			numTokens += tokensPerName
		}
	}
	numTokens += 3 // every reply is primed with <|start|>assistant<|message|>
	return numTokens
}

```


# Available Encodings
 | Encoding name           | OpenAI models                                        |
 | ----------------------- | ---------------------------------------------------- |
 | `o200k_base`            | `gpt-4o`                                             |
 | `cl100k_base`           | `gpt-4`, `gpt-3.5-turbo`, `text-embedding-ada-002`, `text-embedding-3-small`, `text-embedding-3-large`   |
 | `p50k_base`             | Codex models, `text-davinci-002`, `text-davinci-003` |
 | `r50k_base` (or `gpt2`) | GPT-3 models like `davinci`                          |



# Available Models
| Model name                   | OpenAI models |
| ---------------------------- | ------------- |
| gpt-4o-*                     | o200k_base    |
| gpt-4-*                      | cl100k_base   |
| gpt-3.5-turbo-*              | cl100k_base   |
| gpt-4o                       | o200k_base    |
| gpt-4                        | cl100k_base   |
| gpt-3.5-turbo                | cl100k_base   |
| text-davinci-003             | p50k_base     |
| text-davinci-002             | p50k_base     |
| text-davinci-001             | r50k_base     |
| text-curie-001               | r50k_base     |
| text-babbage-001             | r50k_base     |
| text-ada-001                 | r50k_base     |
| davinci                      | r50k_base     |
| curie                        | r50k_base     |
| babbage                      | r50k_base     |
| ada                          | r50k_base     |
| code-davinci-002             | p50k_base     |
| code-davinci-001             | p50k_base     |
| code-cushman-002             | p50k_base     |
| code-cushman-001             | p50k_base     |
| davinci-codex                | p50k_base     |
| cushman-codex                | p50k_base     |
| text-davinci-edit-001        | p50k_edit     |
| code-davinci-edit-001        | p50k_edit     |
| text-embedding-ada-002       | cl100k_base   |
| text-embedding-3-small       | cl100k_base   |
| text-embedding-3-large       | cl100k_base   |
| text-similarity-davinci-001  | r50k_base     |
| text-similarity-curie-001    | r50k_base     |
| text-similarity-babbage-001  | r50k_base     |
| text-similarity-ada-001      | r50k_base     |
| text-search-davinci-doc-001  | r50k_base     |
| text-search-curie-doc-001    | r50k_base     |
| text-search-babbage-doc-001  | r50k_base     |
| text-search-ada-doc-001      | r50k_base     |
| code-search-babbage-code-001 | r50k_base     |
| code-search-ada-code-001     | r50k_base     |
| gpt2                         | gpt2          |



# Test
> you can run test in [test](./test) folder

## compare with original [tiktoken](https://github.com/openai/tiktoken)

## get token by encoding
[result](./doc/test_result.md#encoding-test-result)

## get token by model  
[result](./doc/test_result.md#model-test-result)



# Benchmark
> you can run benchmark in [test](./test) folder

## Benchmark result
| name        | time/op | os         | cpu      | text                             | times  |
| ----------- | ------- | ---------- | -------- | -------------------------------- | ------ |
| tiktoken-go | 8795ns  | macOS 13.2 | Apple M1 | [UDHR](https://unicode.org/udhr) | 100000 |
| tiktoken    | 8838ns  | macOS 13.2 | Apple M1 | [UDHR](https://unicode.org/udhr) | 100000 |

It looks like the performance is almost the same.   

Maybe the difference is due to the difference in the performance of the machine.

Or maybe my benchmark method is not appropriate.  

If you have better benchmark method or if you want add your benchmark result, please feel free to submit a PR.

For new `o200k_base` encoding, it seems slower than `cl100k_base`. tiktoken-go is slightly slower than tiktoken on the following benchmark.

| name        | encoding | time/op | os         | cpu      | text                             | times  |
| ----------- | ------- | ------- | ---------- | -------- | -------------------------------- | ------ |
| tiktoken-go | o200k_base | 108522 ns  | Ubuntu 22.04 | AMD Ryzen 9 5900HS | [UDHR](http://research.ics.aalto.fi/cog/data/udhr/) | 100000 |
| tiktoken    | o200k_base | 70198 ns  | Ubuntu 22.04 | AMD Ryzen 9 5900HS | [UDHR](http://research.ics.aalto.fi/cog/data/udhr/) | 100000 |
| tiktoken-go | cl100k_base | 94502 ns  | Ubuntu 22.04 | AMD Ryzen 9 5900HS | [UDHR](http://research.ics.aalto.fi/cog/data/udhr/) | 100000 |
| tiktoken    | cl100k_base | 54642 ns  | Ubuntu 22.04 | AMD Ryzen 9 5900HS | [UDHR](http://research.ics.aalto.fi/cog/data/udhr/) | 100000 |

# License
[MIT](./LICENSE)
