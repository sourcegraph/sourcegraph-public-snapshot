// Copyright 2023 Mostyn Bramley-Moore.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package github.com/mostynb/go-grpc-compression/nonclobbering/zstd is a
// wrapper for using github.com/klauspost/compress/zstd with gRPC.
//
// If you import this package, it will only register itself as the encoder
// for the "zstd" compressor if no other compressors have already been
// registered with that name.
//
// If you do want to override previously registered "zstd" compressors,
// then you should instead import
// github.com/mostynb/go-grpc-compression/zstd
package zstd

import (
	internalzstd "github.com/mostynb/go-grpc-compression/internal/zstd"

	"github.com/klauspost/compress/zstd"
)

const Name = internalzstd.Name

func init() {
	clobbering := false
	internalzstd.PretendInit(clobbering)
}

var ErrNotInUse = internalzstd.ErrNotInUse

// SetLevel updates the registered compressor to use a particular compression
// level. Returns ErrNotInUse if this module isn't registered (because it has
// been overridden by another encoder with the same name), or any error
// returned by zstd.NewWriter(nil, zstd.WithEncoderLevel(level).
//
// NOTE: this function is not threadsafe and must only be called from an init
// function or from the main goroutine before any other goroutines have been
// created.
func SetLevel(level zstd.EncoderLevel) error {
	return internalzstd.SetLevel(level)
}
