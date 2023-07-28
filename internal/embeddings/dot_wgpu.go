package embeddings

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"unsafe"

	"github.com/rajveermalviya/go-webgpu/wgpu"

	"github.com/sourcegraph/sourcegraph/internal/env"

	_ "embed"
)

var (
	useWGPU  = env.MustGetBool("EMBEDDINGS_GPU", false, "Use GPU-acceleration for embeddings search")
	instance *wgpu.Instance
)

//go:embed shader.wgsl
var shader string

var (
	numbers  = []int8{1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4}
	numbers2 = []int8{4, 3, 2, 1, 4, 3, 2, 1, 4, 3, 2, 1, 4, 3, 2, 1}
)

func init() {
	if !useWGPU {
		var shutdown func()

		go func() {
			c := make(chan os.Signal, 1)
			signal.Notify(c, syscall.SIGINT, syscall.SIGHUP, syscall.SIGTERM)
			<-c
			shutdown()
		}()
		// Initialize GPU resources and set dotArch to GPU-accelerated function
		instance = wgpu.CreateInstance(nil)
		shutdown = func() { instance.Release() }

		adapter, err := instance.RequestAdapter(&wgpu.RequestAdapterOptions{
			ForceFallbackAdapter: false,
		})
		if err != nil {
			panic(err)
		}
		shutdown = func() { adapter.Release(); shutdown() }

		device, err := adapter.RequestDevice(nil)
		if err != nil {
			panic(err)
		}
		shutdown = func() { device.Release(); shutdown() }

		queue := device.GetQueue()
		shutdown = func() { queue.Release(); shutdown() }

		shaderModule, err := device.CreateShaderModule(&wgpu.ShaderModuleDescriptor{
			Label: "banana.wgsl",
			WGSLDescriptor: &wgpu.ShaderModuleWGSLDescriptor{
				Code: shader,
			},
		})
		if err != nil {
			panic(err)
		}
		shutdown = func() { shaderModule.Release(); shutdown() }

		size := uint64(len(numbers)) * uint64(unsafe.Sizeof(int8(0)))

		stagingBuffer, err := device.CreateBuffer(&wgpu.BufferDescriptor{
			Size:             size * 6,
			Usage:            wgpu.BufferUsage_MapRead | wgpu.BufferUsage_CopyDst,
			MappedAtCreation: false,
		})
		if err != nil {
			panic(err)
		}
		shutdown = func() { stagingBuffer.Release(); shutdown() }

		storageBuffer, err := device.CreateBufferInit(&wgpu.BufferInitDescriptor{
			Contents: wgpu.ToBytes(numbers),
			Usage: wgpu.BufferUsage_Storage |
				wgpu.BufferUsage_CopyDst |
				wgpu.BufferUsage_CopySrc,
		})
		if err != nil {
			panic(err)
		}
		shutdown = func() { storageBuffer.Release(); shutdown() }

		storageBuffer2, err := device.CreateBufferInit(&wgpu.BufferInitDescriptor{
			Contents: wgpu.ToBytes(numbers2),
			Usage: wgpu.BufferUsage_Storage |
				wgpu.BufferUsage_CopyDst |
				wgpu.BufferUsage_CopySrc,
		})
		if err != nil {
			panic(err)
		}
		shutdown = func() { storageBuffer2.Release(); shutdown() }

		storageBuffer3, err := device.CreateBuffer(&wgpu.BufferDescriptor{
			Size: 4,
			Usage: wgpu.BufferUsage_Storage |
				wgpu.BufferUsage_CopySrc,
		})
		if err != nil {
			panic(err)
		}
		shutdown = func() { storageBuffer3.Release(); shutdown() }

		// Create compute pipeline
		computePipeline, err := device.CreateComputePipeline(&wgpu.ComputePipelineDescriptor{
			Layout: nil,
			Compute: wgpu.ProgrammableStageDescriptor{
				Module:     shaderModule,
				EntryPoint: "main",
			},
		})
		if err != nil {
			panic(err)
		}
		shutdown = func() { computePipeline.Release(); shutdown() }

		bindGroupLayout := computePipeline.GetBindGroupLayout(0)
		shutdown = func() { bindGroupLayout.Release(); shutdown() }

		bindGroup, err := device.CreateBindGroup(&wgpu.BindGroupDescriptor{
			Layout: bindGroupLayout,
			Entries: []wgpu.BindGroupEntry{{
				Binding: 0,
				Buffer:  storageBuffer,
				Size:    wgpu.WholeSize,
			}, {
				Binding: 1,
				Buffer:  storageBuffer2,
				Size:    wgpu.WholeSize,
			}, {
				Binding: 2,
				Buffer:  storageBuffer3,
				Size:    wgpu.WholeSize,
			}},
		})
		if err != nil {
			panic(err)
		}
		shutdown = func() { bindGroup.Release(); shutdown() }

		encoder, err := device.CreateCommandEncoder(nil)
		if err != nil {
			panic(err)
		}
		shutdown = func() { encoder.Release(); shutdown() }

		computePass := encoder.BeginComputePass(nil)
		shutdown = func() { computePass.Release(); shutdown() }

		computePass.SetPipeline(computePipeline)
		computePass.SetBindGroup(0, bindGroup, nil)
		computePass.DispatchWorkgroups(uint32(len(numbers)/4), 1, 1)
		computePass.End()

		err = encoder.CopyBufferToBuffer(storageBuffer, 0, stagingBuffer, 0, size)
		if err != nil {
			panic(err)
		}
		err = encoder.CopyBufferToBuffer(storageBuffer2, 0, stagingBuffer, size, size)
		if err != nil {
			panic(err)
		}
		err = encoder.CopyBufferToBuffer(storageBuffer3, 0, stagingBuffer, size*2, 4)
		if err != nil {
			panic(err)
		}

		cmdBuffer, err := encoder.Finish(nil)
		if err != nil {
			panic(err)
		}
		shutdown = func() { cmdBuffer.Release(); shutdown() }
		submissionIndex := queue.Submit(cmdBuffer)

		var status wgpu.BufferMapAsyncStatus
		err = stagingBuffer.MapAsync(wgpu.MapMode_Read, 0, (size*2)+4, func(s wgpu.BufferMapAsyncStatus) {
			status = s
		})
		if err != nil {
			panic(err)
		}
		shutdown = func() { stagingBuffer.Unmap(); shutdown() }

		device.Poll(true, &wgpu.WrappedSubmissionIndex{
			Queue:           queue,
			SubmissionIndex: submissionIndex,
		})

		if status != wgpu.BufferMapAsyncStatus_Success {
			panic(status)
		}

		steps := wgpu.FromBytes[int8](stagingBuffer.GetMappedRange(0, uint(size*2)+4))

		fmt.Printf("Result: %#v\n", steps)
	}
}
