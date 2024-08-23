# tiktoken-go
Go 语言版本的 OpenAI 的 tiktoken。  
帮你把文本转换成 OpenAI 的模型可以识别的 token。  
tiktoken的原项目地址[tiktoken](https://github.com/openai/tiktoken).  

# 用法

## 安装

```bash
go get github.com/pkoukk/tiktoken-go
```
## 缓存
Tiktoken-go 和原始的 Tiktoken 库一样，具有相同的缓存机制。  

您可以使用环境变量 TIKTOKEN_CACHE_DIR 来设置缓存目录。  

一旦设置了该变量，tiktoken-go 将使用该目录来缓存令牌字典。  

如果您未设置此环境变量，则 tiktoken-go 将在每次首次初始化编码时下载字典。  


## 替代 BPE 加载器
默认情况下，tiktoken-go 会在运行时下载字典，如果您不想使用缓存或每次下载字典，您可以使用替代 BPE 加载器。

只需在调用 `tiktoken.GetEncoding` 或 `tiktoken.EncodingForModel` 之前调用 `tiktoken.SetBpeLoader`。

`BpeLoader` 是一个接口，您可以通过实现此接口来实现自己的 BPE 加载器。

### 离线 BPE 加载器
离线 BPE 加载器从嵌入文件加载 BPE 字典。

由于 BPE 字典的文件较大，不适合包含在本项目中，故此加载器在其他项目中。

如果需要使用，请引用：[tiktoken_loader](https://github.com/pkoukk/tiktoken-go-loader)

## 例子

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

	// 如果你不想在运行时下载字典，你可以使用离线加载器
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

### get token by Model

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

### 计算chat API消息当中的token消耗
这段代码根据[官方示例](https://github.com/openai/openai-cookbook/blob/main/examples/How_to_count_tokens_with_tiktoken.ipynb)编写

编写时间： `2023-06-28`

请注意，消息的token计算方式可能随时会发生改变，以下代码并不一定在将来适用，如果您需要精确的计算，请关注官方文档。

如果您发现这段代码不再适用，欢迎您提PR或Issue。

```go
package main

import (
	"fmt"

	"github.com/pkoukk/tiktoken-go"
	"github.com/sashabaranov/go-openai"
)

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

# available encodings
 | Encoding name           | OpenAI models                                        |
 | ----------------------- | ---------------------------------------------------- |
 | `o200k_base`            | `gpt-4o`                                             |
 | `cl100k_base`           | `gpt-4`, `gpt-3.5-turbo`, `text-embedding-ada-002`, 	`text-embedding-3-small`, `text-embedding-3-large`   |
 | `p50k_base`             | Codex models, `text-davinci-002`, `text-davinci-003` |
 | `r50k_base` (or `gpt2`) | GPT-3 models like `davinci`                          |


# available models
| Model name                   | OpenAI models |
| ---------------------------- | ------------- |
| gpt-4o-*                     | o200k_base    |
| gpt-4                        | cl100k_base   |
| gpt-4-*                      | cl100k_base   |
| gpt-3.5-turbo                | cl100k_base   |
| gpt-3.5-turbo-*              | cl100k_base   |
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

# 与官方 [tiktoken](https://github.com/openai/tiktoken) 的对比

## get token by encoding
[测试结果](./doc/test_result.md#encoding-test-result)

## get token by model  
[测试结果](./doc/test_result.md#model-test-result)

# Benchmark
> 你可以使用 [test](./test) 目录下的文件执行基准测试。 

## Benchmark result
| name        | time/op | os         | cpu      | text                             | times  |
| ----------- | ------- | ---------- | -------- | -------------------------------- | ------ |
| tiktoken-go | 8795ns  | macOS 13.2 | Apple M1 | [UDHR](https://unicode.org/udhr) | 100000 |
| tiktoken    | 8838ns  | macOS 13.2 | Apple M1 | [UDHR](https://unicode.org/udhr) | 100000 |

看上去tiktoken-go的性能基本与原tiktoken一致。  

也许在不同的机器上的测试结果会有所不同。也可能是我的测试方法并不恰当。

如果你有更好的测试方法，或者说你想添加在你机器上的测试结果，欢迎提PR。

新的 `o200k_base` 编码, 看起来比 `cl100k_base` 慢. 在以下硬件上，tiktoken-go 比 tiktoken 略慢。

| name        | encoding | time/op | os         | cpu      | text                             | times  |
| ----------- | ------- | ------- | ---------- | -------- | -------------------------------- | ------ |
| tiktoken-go | o200k_base | 108522 ns  | Ubuntu 22.04 | AMD Ryzen 9 5900HS | [UDHR](http://research.ics.aalto.fi/cog/data/udhr/) | 100000 |
| tiktoken    | o200k_base | 70198 ns  | Ubuntu 22.04 | AMD Ryzen 9 5900HS | [UDHR](http://research.ics.aalto.fi/cog/data/udhr/) | 100000 |
| tiktoken-go | cl100k_base | 94502 ns  | Ubuntu 22.04 | AMD Ryzen 9 5900HS | [UDHR](http://research.ics.aalto.fi/cog/data/udhr/) | 100000 |
| tiktoken    | cl100k_base | 54642 ns  | Ubuntu 22.04 | AMD Ryzen 9 5900HS | [UDHR](http://research.ics.aalto.fi/cog/data/udhr/) | 100000 |

# License
[MIT](./LICENSE)
