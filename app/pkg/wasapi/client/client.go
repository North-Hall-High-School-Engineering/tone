package client

import (
	"app/pkg/wasapi/com"
	"fmt"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

type AUDCLNT_SHAREMODE uint32

const (
	AUDCLNT_SHAREMODE_SHARED AUDCLNT_SHAREMODE = iota
	AUDCLNT_SHAREMODE_EXCLUSIVE
)

const (
	AUDCLNT_STREAMFLAGS_CROSSPROCESS        = 0x00010000
	AUDCLNT_STREAMFLAGS_LOOPBACK            = 0x00020000
	AUDCLNT_STREAMFLAGS_EVENTCALLBACK       = 0x00040000
	AUDCLNT_STREAMFLAGS_NOPERSIST           = 0x00080000
	AUDCLNT_STREAMFLAGS_RATEADJUST          = 0x00100000
	AUDCLNT_STREAMFLAGS_AUTOCONVERTPCM      = 0x80000000
	AUDCLNT_STREAMFLAGS_SRC_DEFAULT_QUALITY = 0x08000000
)

const (
	AUDCLNT_BUFFERFLAGS_DATA_DISCONTINUITY uint32 = iota
	AUDCLNT_BUFFERFLAGS_SILENT
	AUDCLNT_BUFFERFLAGS_TIMESTAMP_ERROR
)

var _IID_IAudioClient = windows.GUID{Data1: 0x1CB9AD4C, Data2: 0xDBFA, Data3: 0x4C32, Data4: [8]byte{0xB1, 0x78, 0xC2, 0xF5, 0x68, 0xA7, 0x03, 0xB2}}

func IID_IAudioClient() windows.GUID {
	return _IID_IAudioClient
}

type IAudioClient struct {
	vtbl *_IAudioClientVtbl
}

type _IAudioClientVtbl struct {
	com.IUnknownVtbl
	Initialize        uintptr
	GetBufferSize     uintptr
	GetStreamLatency  uintptr
	GetCurrentPadding uintptr
	IsFormatSupported uintptr
	GetMixFormat      uintptr
	GetDevicePeriod   uintptr
	Start             uintptr
	Stop              uintptr
	Reset             uintptr
	SetEventHandle    uintptr
	GetService        uintptr
}

func (client *IAudioClient) Release() (err error) {
	r, _, _ := syscall.SyscallN(client.vtbl.Release, uintptr(unsafe.Pointer(client)))

	if com.HRESULT(r) != com.HRESULT(windows.S_OK) {
		err = fmt.Errorf("IAudioClient::Release failed with code: 0x%08X", com.HRESULT(r))
		return
	}

	return
}

func (client *IAudioClient) Initialize(
	shareMode AUDCLNT_SHAREMODE,
	streamFlags uint32,
	hnsBufferDuration uint64,
	hnsPeriodicity uint64,
	format *WAVEFORMATEXTENSIBLE,
	audioSessionGuid *windows.GUID,
) (err error) {
	formatBuf := format.toBytes()
	r, _, _ := syscall.SyscallN(client.vtbl.Initialize, uintptr(unsafe.Pointer(client)),
		uintptr(shareMode),
		uintptr(streamFlags),
		uintptr(hnsBufferDuration),
		uintptr(hnsPeriodicity),
		uintptr(unsafe.Pointer(&formatBuf[0])),
		uintptr(unsafe.Pointer(audioSessionGuid)),
	)

	if com.HRESULT(r) != com.HRESULT(windows.S_OK) {
		err = fmt.Errorf("IAudioClient::Initialize failed with code: 0x%08X", com.HRESULT(r))
		return
	}

	return
}

func (client *IAudioClient) Start() (err error) {
	r, _, _ := syscall.SyscallN(client.vtbl.Start, uintptr(unsafe.Pointer(client)))

	if com.HRESULT(r) != com.HRESULT(windows.S_OK) {
		err = fmt.Errorf("IAudioClient::Start failed with code: 0x%08X", com.HRESULT(r))
		return
	}

	return
}

func (client *IAudioClient) Stop() (err error) {
	r, _, _ := syscall.SyscallN(client.vtbl.Stop, uintptr(unsafe.Pointer(client)))

	if com.HRESULT(r) != com.HRESULT(windows.S_OK) {
		err = fmt.Errorf("IAudioClient::Stop failed with code: 0x%08X", com.HRESULT(r))
		return
	}

	return
}

func (client *IAudioClient) Reset() (err error) {
	r, _, _ := syscall.SyscallN(client.vtbl.Reset, uintptr(unsafe.Pointer(client)))

	if com.HRESULT(r) != com.HRESULT(windows.S_OK) {
		err = fmt.Errorf("IAudioClient::Reset failed with code: 0x%08X", com.HRESULT(r))
		return
	}

	return
}

func (client *IAudioClient) GetMixFormat() (deviceFormat WAVEFORMATEXTENSIBLE, err error) {
	var formatPtr *byte
	r, _, _ := syscall.SyscallN(client.vtbl.GetMixFormat, uintptr(unsafe.Pointer(client)),
		uintptr(unsafe.Pointer(&formatPtr)),
	)

	if com.HRESULT(r) != com.HRESULT(windows.S_OK) {
		err = fmt.Errorf("IAudioClient::GetMixFormat failed with code: 0x%08X", com.HRESULT(r))
		return
	}

	buf := unsafe.Slice(formatPtr, 40)
	deviceFormat.fromBytes(buf)

	// 释放内存
	com.CoTaskMemFree(unsafe.Pointer(formatPtr))

	return
}

func (client *IAudioClient) GetBufferSize() (numBufferFrames uint32, err error) {
	r, _, _ := syscall.SyscallN(client.vtbl.GetBufferSize, uintptr(unsafe.Pointer(client)),
		uintptr(unsafe.Pointer(&numBufferFrames)),
	)

	if com.HRESULT(r) != com.HRESULT(windows.S_OK) {
		err = fmt.Errorf("IAudioClient::GetBufferSize failed with code: 0x%08X", com.HRESULT(r))
		return
	}

	return
}

func (client *IAudioClient) GetService(iid *windows.GUID) (v unsafe.Pointer, err error) {
	r, _, _ := syscall.SyscallN(client.vtbl.GetService, uintptr(unsafe.Pointer(client)),
		uintptr(unsafe.Pointer(iid)),
		uintptr(unsafe.Pointer(&v)),
	)

	if com.HRESULT(r) != com.HRESULT(windows.S_OK) {
		err = fmt.Errorf("IAudioClient::GetService failed with code: 0x%08X", com.HRESULT(r))
		return
	}

	return
}
