package mmdevice

import (
	"app/pkg/wasapi/com"
	"fmt"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

var _IID_IMMDeviceEnumerator = windows.GUID{Data1: 0xA95664D2, Data2: 0x9614, Data3: 0x4F35, Data4: [8]byte{0xA7, 0x46, 0xDE, 0x8D, 0xB6, 0x36, 0x17, 0xE6}}

func IID_IMMDeviceEnumerator() windows.GUID {
	return _IID_IMMDeviceEnumerator
}

type IMMDeviceEnumerator struct {
	vtbl *_IMMDeviceEnumeratorVtbl
}

type _IMMDeviceEnumeratorVtbl struct {
	com.IUnknownVtbl
	EnumAudioEndpoints                     uintptr
	GetDefaultAudioEndpoint                uintptr
	GetDevice                              uintptr
	RegisterEndpointNotificationCallback   uintptr
	UnregisterEndpointNotificationCallback uintptr
}

func (enumerator *IMMDeviceEnumerator) Release() (err error) {
	r, _, _ := syscall.SyscallN(enumerator.vtbl.Release, uintptr(unsafe.Pointer(enumerator)))

	if com.HRESULT(r) != com.HRESULT(windows.S_OK) {
		err = fmt.Errorf("IMMDeviceEnumerator::Release failed with code: 0x%08X", com.HRESULT(r))
		return
	}

	return
}

func (enumerator *IMMDeviceEnumerator) GetDefaultAudioEndpoint(dataFlow EDataFlow, role ERole) (endpoint *IMMDevice, err error) {
	r, _, _ := syscall.SyscallN(enumerator.vtbl.GetDefaultAudioEndpoint, uintptr(unsafe.Pointer(enumerator)),
		uintptr(dataFlow),
		uintptr(role),
		uintptr(unsafe.Pointer(&endpoint)),
	)

	if com.HRESULT(r) != com.HRESULT(windows.S_OK) {
		err = fmt.Errorf("IMMDeviceEnumerator::GetDefaultAudioEndpoint failed with code: 0x%08X", com.HRESULT(r))
		return
	}

	return
}

func (enumerator *IMMDeviceEnumerator) GetDevice(id string) (device *IMMDevice, err error) {
	var utf16ptr *uint16

	if utf16ptr, err = windows.UTF16PtrFromString(id); err != nil {
		return
	}

	r, _, _ := syscall.SyscallN(enumerator.vtbl.GetDevice, uintptr(unsafe.Pointer(enumerator)),
		uintptr(unsafe.Pointer(utf16ptr)),
		uintptr(unsafe.Pointer(&device)),
	)

	if com.HRESULT(r) != com.HRESULT(windows.S_OK) {
		err = fmt.Errorf("IMMDeviceEnumerator::GetDevice failed with code: 0x%08X", com.HRESULT(r))
		return
	}

	return
}
