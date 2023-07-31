package embeddings

import (
	"fmt"
	"math"
	"os"
	"os/signal"
	"syscall"

	"github.com/rajveermalviya/go-webgpu/wgpu"

	"github.com/sourcegraph/sourcegraph/internal/env"

	_ "embed"
)

var useWGPU = env.MustGetBool("EMBEDDINGS_GPU", false, "Use GPU-acceleration for embeddings search")

//go:embed shader.wgsl
var shader string

var (
	wgpuDevice          *wgpu.Device
	wgpuComputePipeline *wgpu.ComputePipeline
	wgpuQueue           *wgpu.Queue
)

const (
	int32Size     = 4
	workGroupSize = 1024
)

func init() {
	if useWGPU {
		var shutdown func()

		go func() {
			c := make(chan os.Signal, 1)
			signal.Notify(c, syscall.SIGINT, syscall.SIGHUP, syscall.SIGTERM)
			<-c
			shutdown()
		}()
		// Initialize GPU resources and set dotArch to GPU-accelerated function
		instance := wgpu.CreateInstance(nil)
		shutdown = func() { instance.Release() }

		adapter, err := instance.RequestAdapter(&wgpu.RequestAdapterOptions{ForceFallbackAdapter: false})
		if err != nil {
			panic(err)
		}
		shutdown = func() { adapter.Release(); shutdown() }

		limits := wgpu.DefaultLimits()
		limits.MaxComputeWorkgroupSizeX = workGroupSize
		limits.MaxComputeInvocationsPerWorkgroup = workGroupSize
		wgpuDevice, err = adapter.RequestDevice(&wgpu.DeviceDescriptor{
			RequiredLimits: &wgpu.RequiredLimits{
				Limits: limits,
			},
		})
		if err != nil {
			panic(err)
		}
		shutdown = func() { wgpuDevice.Release(); shutdown() }

		wgpuQueue = wgpuDevice.GetQueue()
		shutdown = func() { wgpuQueue.Release(); shutdown() }

		shaderModule, err := wgpuDevice.CreateShaderModule(&wgpu.ShaderModuleDescriptor{
			Label: "banana.wgsl",
			WGSLDescriptor: &wgpu.ShaderModuleWGSLDescriptor{
				Code: shader,
			},
		})
		if err != nil {
			panic(err)
		}
		shutdown = func() { shaderModule.Release(); shutdown() }

		// Create compute pipeline
		wgpuComputePipeline, err = wgpuDevice.CreateComputePipeline(&wgpu.ComputePipelineDescriptor{
			Layout: nil,
			Compute: wgpu.ProgrammableStageDescriptor{
				Module:     shaderModule,
				EntryPoint: "main",
			},
		})
		if err != nil {
			panic(err)
		}
		shutdown = func() { wgpuComputePipeline.Release(); shutdown() }

		dotArch = dotWGPU
	}
}

func dotWGPU(a, b []int8) int32 {
	if len(a) < 4 {
		a = append(a, 0, 0, 0, 0)
	}
	if len(b) < 4 {
		b = append(b, 0, 0, 0, 0)
	}

	// this should always be a whole number, assuming the inputs are always of a
	// lngth divisible by 4
	gpuSideArraySize := len(a) / 4
	numWorkGroups := int(math.Ceil(float64(gpuSideArraySize) / float64(workGroupSize)))

	storageBuffer, err := wgpuDevice.CreateBufferInit(&wgpu.BufferInitDescriptor{
		Contents: wgpu.ToBytes(a),
		Usage:    wgpu.BufferUsage_Storage | wgpu.BufferUsage_CopySrc,
	})
	if err != nil {
		panic(err)
	}
	defer storageBuffer.Release()

	storageBuffer2, err := wgpuDevice.CreateBufferInit(&wgpu.BufferInitDescriptor{
		Contents: wgpu.ToBytes(b),
		Usage:    wgpu.BufferUsage_Storage | wgpu.BufferUsage_CopySrc,
	})
	if err != nil {
		panic(err)
	}
	defer storageBuffer2.Release()

	storageBuffer3, err := wgpuDevice.CreateBuffer(&wgpu.BufferDescriptor{
		Size:  workGroupSize * int32Size,
		Usage: wgpu.BufferUsage_Storage | wgpu.BufferUsage_CopySrc,
	})
	if err != nil {
		panic(err)
	}
	defer storageBuffer3.Release()

	// unpaddedSize := uint64(len(a)*int(unsafe.Sizeof(int8(0))))*2 + 4
	paddedSize := storageBuffer.GetSize() + storageBuffer2.GetSize() + storageBuffer3.GetSize()

	stagingBuffer, err := wgpuDevice.CreateBuffer(&wgpu.BufferDescriptor{
		Size:             paddedSize,
		Usage:            wgpu.BufferUsage_MapRead | wgpu.BufferUsage_CopyDst,
		MappedAtCreation: false,
	})
	if err != nil {
		panic(err)
	}
	defer stagingBuffer.Release()

	bindGroupLayout := wgpuComputePipeline.GetBindGroupLayout(0)
	defer bindGroupLayout.Release()

	bindGroup, err := wgpuDevice.CreateBindGroup(&wgpu.BindGroupDescriptor{
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
	defer bindGroup.Release()

	encoder, err := wgpuDevice.CreateCommandEncoder(nil)
	if err != nil {
		panic(err)
	}
	defer encoder.Release()

	computePass := encoder.BeginComputePass(nil)
	defer computePass.Release()

	computePass.SetPipeline(wgpuComputePipeline)
	computePass.SetBindGroup(0, bindGroup, nil)
	computePass.DispatchWorkgroups(uint32(numWorkGroups), 1, 1)
	if err := computePass.End(); err != nil {
		panic(err)
	}

	err = encoder.CopyBufferToBuffer(storageBuffer, 0, stagingBuffer, 0, storageBuffer.GetSize())
	if err != nil {
		panic(err)
	}
	err = encoder.CopyBufferToBuffer(storageBuffer2, 0, stagingBuffer, storageBuffer.GetSize(), storageBuffer2.GetSize())
	if err != nil {
		panic(err)
	}
	err = encoder.CopyBufferToBuffer(storageBuffer3, 0, stagingBuffer, storageBuffer.GetSize()+storageBuffer2.GetSize(), uint64(numWorkGroups)*int32Size)
	if err != nil {
		panic(err)
	}

	cmdBuffer, err := encoder.Finish(nil)
	if err != nil {
		panic(err)
	}
	defer cmdBuffer.Release()
	submissionIndex := wgpuQueue.Submit(cmdBuffer)

	var status wgpu.BufferMapAsyncStatus
	err = stagingBuffer.MapAsync(wgpu.MapMode_Read, 0, paddedSize, func(s wgpu.BufferMapAsyncStatus) {
		status = s
	})
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := stagingBuffer.Unmap(); err != nil {
			fmt.Println(err)
		}
	}()

	wgpuDevice.Poll(true, &wgpu.WrappedSubmissionIndex{
		Queue:           wgpuQueue,
		SubmissionIndex: submissionIndex,
	})

	if status != wgpu.BufferMapAsyncStatus_Success {
		panic(status)
	}

	result := wgpu.FromBytes[int32](stagingBuffer.GetMappedRange(uint(storageBuffer.GetSize()+storageBuffer2.GetSize()), uint(numWorkGroups)*int32Size))
	// steps := wgpu.FromBytes[int8](stagingBuffer.GetMappedRange(0, uint(unpaddedSize)))

	// fmt.Printf("Result: %#v\n", steps)

	// end := steps[len(steps)-4:]
	// num := binary.LittleEndian.Uint32(*(*[]byte)(unsafe.Pointer(&end)))

	// probably cheaper to sum the partials CPU side instead of firing off another dispatch
	var sum int32
	for i := 0; i < numWorkGroups; i++ {
		sum += result[i]
	}

	// fmt.Println(num, len(end))
	// return steps[0]
	// return int32(steps[len(steps)-5])
	return sum
}
