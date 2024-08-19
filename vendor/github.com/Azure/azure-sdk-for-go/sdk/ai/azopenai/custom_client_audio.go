//go:build go1.18
// +build go1.18

// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.

package azopenai

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/streaming"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
)

// GetAudioTranscriptionOptions contains the optional parameters for the [Client.GetAudioTranscription] method.
type GetAudioTranscriptionOptions struct {
	// placeholder for future optional parameters
}

// GetAudioTranscriptionResponse contains the response from method [Client.GetAudioTranscription].
type GetAudioTranscriptionResponse struct {
	AudioTranscription
}

// GetAudioTranscription gets transcribed text and associated metadata from provided spoken audio data. Audio will
// be transcribed in the written language corresponding to the language it was spoken in. Gets transcribed text
// and associated metadata from provided spoken audio data. Audio will be transcribed in the written language corresponding
// to the language it was spoken in.
// If the operation fails it returns an *azcore.ResponseError type.
//
// Generated from API version 2023-09-01-preview
//   - body - contains parameters to specify audio data to transcribe and control the transcription.
//   - options - optional parameters for this method.
func (client *Client) GetAudioTranscription(ctx context.Context, body AudioTranscriptionOptions, options *GetAudioTranscriptionOptions) (GetAudioTranscriptionResponse, error) {
	resp, err := client.getAudioTranscriptionInternal(ctx, streaming.NopCloser(bytes.NewReader(body.File)), &getAudioTranscriptionInternalOptions{
		Filename:       body.Filename,
		Language:       body.Language,
		DeploymentName: body.DeploymentName,
		Prompt:         body.Prompt,
		ResponseFormat: body.ResponseFormat,
		Temperature:    body.Temperature,
	})

	if err != nil {
		return GetAudioTranscriptionResponse{}, err
	}

	return GetAudioTranscriptionResponse(resp), nil
}

// GetAudioTranslationOptions contains the optional parameters for the [Client.GetAudioTranslation] method.
type GetAudioTranslationOptions struct {
	// placeholder for future optional parameters
}

// GetAudioTranslationResponse contains the response from method [Client.GetAudioTranslation].
type GetAudioTranslationResponse struct {
	AudioTranslation
}

// GetAudioTranslation gets English language transcribed text and associated metadata from provided spoken audio
// data. Gets English language transcribed text and associated metadata from provided spoken audio data.
// If the operation fails it returns an *azcore.ResponseError type.
//
// Generated from API version 2023-09-01-preview
//   - body - contains parameters to specify audio data to translate and control the translation.
//   - options - optional parameters for this method.
func (client *Client) GetAudioTranslation(ctx context.Context, body AudioTranslationOptions, options *GetAudioTranslationOptions) (GetAudioTranslationResponse, error) {
	resp, err := client.getAudioTranslationInternal(ctx, streaming.NopCloser(bytes.NewReader(body.File)), &getAudioTranslationInternalOptions{
		Filename:       body.Filename,
		DeploymentName: body.DeploymentName,
		Prompt:         body.Prompt,
		ResponseFormat: body.ResponseFormat,
		Temperature:    body.Temperature,
	})

	if err != nil {
		return GetAudioTranslationResponse{}, err
	}

	return GetAudioTranslationResponse(resp), nil
}

func setMultipartFormData[T getAudioTranscriptionInternalOptions | getAudioTranslationInternalOptions](req *policy.Request, file io.ReadSeekCloser, options T) error {
	body := bytes.Buffer{}
	writer := multipart.NewWriter(&body)

	writeContent := func(fieldname, filename string, file io.ReadSeekCloser) error {
		fd, err := writer.CreateFormFile(fieldname, filename)

		if err != nil {
			return err
		}

		if _, err := io.Copy(fd, file); err != nil {
			return err
		}

		return err
	}

	var filename = "audio.mp3"

	switch opt := any(options).(type) {
	case getAudioTranscriptionInternalOptions:
		if opt.Filename != nil {
			filename = *opt.Filename
		}
	case getAudioTranslationInternalOptions:
		if opt.Filename != nil {
			filename = *opt.Filename
		}
	}

	if err := writeContent("file", filename, file); err != nil {
		return err
	}

	switch v := any(options).(type) {
	case getAudioTranslationInternalOptions:
		if err := writeField(writer, "model", v.DeploymentName); err != nil {
			return err
		}
		if err := writeField(writer, "prompt", v.Prompt); err != nil {
			return err
		}
		if err := writeField(writer, "response_format", v.ResponseFormat); err != nil {
			return err
		}
		if err := writeField(writer, "temperature", v.Temperature); err != nil {
			return err
		}
	case getAudioTranscriptionInternalOptions:
		if err := writeField(writer, "language", v.Language); err != nil {
			return err
		}
		if err := writeField(writer, "model", v.DeploymentName); err != nil {
			return err
		}
		if err := writeField(writer, "prompt", v.Prompt); err != nil {
			return err
		}
		if err := writeField(writer, "response_format", v.ResponseFormat); err != nil {
			return err
		}
		if err := writeField(writer, "temperature", v.Temperature); err != nil {
			return err
		}
	default:
		return fmt.Errorf("failed to serialize multipart for unhandled type %T", body)
	}

	if err := writer.Close(); err != nil {
		return err
	}

	return req.SetBody(streaming.NopCloser(bytes.NewReader(body.Bytes())), writer.FormDataContentType())
}

func getAudioTranscriptionInternalHandleResponse(resp *http.Response) (getAudioTranscriptionInternalResponse, error) {
	at, err := deserializeAudioTranscription(resp)

	if err != nil {
		return getAudioTranscriptionInternalResponse{}, err
	}

	return getAudioTranscriptionInternalResponse{AudioTranscription: at}, nil
}

func getAudioTranslationInternalHandleResponse(resp *http.Response) (getAudioTranslationInternalResponse, error) {
	at, err := deserializeAudioTranslation(resp)

	if err != nil {
		return getAudioTranslationInternalResponse{}, err
	}

	return getAudioTranslationInternalResponse{AudioTranslation: at}, nil
}

// deserializeAudioTranscription handles deserializing the content if it's text/plain
// or a JSON object.
func deserializeAudioTranscription(resp *http.Response) (AudioTranscription, error) {
	defer func() {
		_ = resp.Request.Body.Close()
	}()

	contentType := resp.Header.Get("Content-type")

	if strings.Contains(contentType, "text/plain") {
		body, err := io.ReadAll(resp.Body)

		if err != nil {
			return AudioTranscription{}, err
		}

		return AudioTranscription{
			Text: to.Ptr(string(body)),
		}, nil
	}

	var result *AudioTranscription
	if err := runtime.UnmarshalAsJSON(resp, &result); err != nil {
		return AudioTranscription{}, err
	}

	return *result, nil
}

// deserializeAudioTranslation handles deserializing the content if it's text/plain
// or a JSON object.
func deserializeAudioTranslation(resp *http.Response) (AudioTranslation, error) {
	defer func() {
		_ = resp.Request.Body.Close()
	}()

	contentType := resp.Header.Get("Content-type")

	if strings.Contains(contentType, "text/plain") {
		body, err := io.ReadAll(resp.Body)

		if err != nil {
			return AudioTranslation{}, err
		}

		return AudioTranslation{
			Text: to.Ptr(string(body)),
		}, nil
	}

	var result *AudioTranslation
	if err := runtime.UnmarshalAsJSON(resp, &result); err != nil {
		return AudioTranslation{}, err
	}

	return *result, nil
}

func writeField[T interface {
	string | float32 | AudioTranscriptionFormat | AudioTranslationFormat
}](writer *multipart.Writer, fieldName string, v *T) error {
	if v == nil {
		return nil
	}

	switch v2 := any(v).(type) {
	case *string:
		return writer.WriteField(fieldName, *v2)
	case *float32:
		return writer.WriteField(fieldName, fmt.Sprintf("%f", *v2))
	case *AudioTranscriptionFormat:
		return writer.WriteField(fieldName, string(*v2))
	case *AudioTranslationFormat:
		return writer.WriteField(fieldName, string(*v2))
	default:
		return fmt.Errorf("no handler for type %T", v)
	}
}
