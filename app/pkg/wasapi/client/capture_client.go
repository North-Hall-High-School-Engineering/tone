package client

import (
	"app/pkg/wasapi/com"
	"fmt"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

var _IID_IAudioCaptureClient = windows.GUID{Data1: 0xC8ADBD64, Data2: 0xE71E, Data3: 0x48a0, Data4: [8]byte{0xA4, 0xDE, 0x18, 0x5C, 0x39, 0x5C, 0xD3, 0x17}}

func IID_IAudioCaptureClient() windows.GUID {
	return _IID_IAudioCaptureClient
}

type IAudioCaptureClient struct {
	vtbl *_IAudioCaptureClientVtbl
}

type _IAudioCaptureClientVtbl struct {
	com.IUnknownVtbl
	GetBuffer         uintptr
	ReleaseBuffer     uintptr
	GetNextPacketSize uintptr
}

func (client *IAudioCaptureClient) Release() (err error) {
	r, _, _ := syscall.SyscallN(client.vtbl.Release, uintptr(unsafe.Pointer(client)))

	if com.HRESULT(r) != com.HRESULT(windows.S_OK) {
		err = fmt.Errorf("IAudioClient::Release failed with code: 0x%08X", com.HRESULT(r))
		return
	}

	return
}

func (client *IAudioCaptureClient) GetBuffer() (data []byte, numFramesToRead uint32, flags uint32, devicePosition uint64, QPCPosition uint64, err error) {
	var buf *uint8
	r, _, _ := syscall.SyscallN(client.vtbl.GetBuffer, uintptr(unsafe.Pointer(client)),
		uintptr(unsafe.Pointer(&buf)),
		uintptr(unsafe.Pointer(&numFramesToRead)),
		uintptr(unsafe.Pointer(&flags)),
		uintptr(unsafe.Pointer(&devicePosition)),
		uintptr(unsafe.Pointer(&QPCPosition)),
	)

	if com.HRESULT(r) != com.HRESULT(windows.S_OK) {
		err = fmt.Errorf("IAudioCaptureClient::GetBuffer failed with code: 0x%08X", com.HRESULT(r))
		return
	}

	data = unsafe.Slice(buf, numFramesToRead)

	return
}

func (client *IAudioCaptureClient) ReleaseBuffer(numFramesToRead uint32) (err error) {
	r, _, _ := syscall.SyscallN(client.vtbl.ReleaseBuffer, uintptr(unsafe.Pointer(client)),
		uintptr(numFramesToRead),
	)

	if com.HRESULT(r) != com.HRESULT(windows.S_OK) {
		err = fmt.Errorf("IAudioCaptureClient::ReleaseBuffer failed with code: 0x%08X", com.HRESULT(r))
		return
	}

	return
}

func (client *IAudioCaptureClient) GetNextPacketSize() (numFramesInNextPacket uint32, err error) {
	r, _, _ := syscall.SyscallN(client.vtbl.GetNextPacketSize, uintptr(unsafe.Pointer(client)),
		uintptr(unsafe.Pointer(&numFramesInNextPacket)),
	)

	if com.HRESULT(r) != com.HRESULT(windows.S_OK) {
		err = fmt.Errorf("IAudioCaptureClient::GetNextPacketSize failed with code: 0x%08X", com.HRESULT(r))
		return
	}

	return
}
