package mmdevice

import (
	"app/pkg/wasapi/com"
	"fmt"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	DEVICE_STATE_ACTIVE     = 0x00000001
	DEVICE_STATE_DISABLED   = 0x00000002
	DEVICE_STATE_NOTPRESENT = 0x00000004
	DEVICE_STATE_UNPLUGGED  = 0x00000008
	DEVICE_STATEMASK_ALL    = 0x0000000F
)

type EDataFlow uint32

const (
	ERender EDataFlow = iota
	ECapture
	EAll
)

type ERole uint32

const (
	EConsole ERole = iota
	EMultimedia
	ECommunications
)

var _CLSID_MMDeviceEnumerator = windows.GUID{Data1: 0xBCDE0395, Data2: 0xE52F, Data3: 0x467C, Data4: [8]byte{0x8E, 0x3D, 0xC4, 0x57, 0x92, 0x91, 0x69, 0x2E}}

func CLSID_MMDeviceEnumerator() windows.GUID {
	return _CLSID_MMDeviceEnumerator
}

var _IID_IMMDevice = windows.GUID{Data1: 0xD666063F, Data2: 0x1587, Data3: 0x4E43, Data4: [8]byte{0x81, 0xF1, 0xB9, 0x48, 0xE8, 0x07, 0x36, 0x3F}}

func IID_IMMDevice() windows.GUID {
	return _IID_IMMDevice
}

type IMMDevice struct {
	vtbl *_IMMDeviceVtbl
}

type _IMMDeviceVtbl struct {
	com.IUnknownVtbl
	Activate          uintptr
	OpenPropertyStore uintptr
	GetId             uintptr
	GetState          uintptr
}

func (device *IMMDevice) Release() (err error) {
	r, _, _ := syscall.SyscallN(device.vtbl.Release, uintptr(unsafe.Pointer(device)))

	if com.HRESULT(r) != com.HRESULT(windows.S_OK) {
		err = fmt.Errorf("IMMDevice::Release failed with code: 0x%08X", com.HRESULT(r))
		return
	}

	return
}

func (device *IMMDevice) Activate(iid windows.GUID, clsCtx uint32 /*activationParams *com.PROPVARIANT*/) (ppInterface unsafe.Pointer, err error) {
	r, _, _ := syscall.SyscallN(device.vtbl.Activate, uintptr(unsafe.Pointer(device)),
		uintptr(unsafe.Pointer(&iid)),
		uintptr(clsCtx),
		/*uintptr(unsafe.Pointer(activationParams))*/ 0,
		uintptr(unsafe.Pointer(&ppInterface)),
	)

	if com.HRESULT(r) != com.HRESULT(windows.S_OK) {
		err = fmt.Errorf("IMMDevice::Activate failed with code: 0x%08X", com.HRESULT(r))
		return
	}

	return
}

func (device *IMMDevice) GetId() (id string, err error) {
	var utf16ptr *uint16
	r, _, _ := syscall.SyscallN(device.vtbl.GetId, uintptr(unsafe.Pointer(device)),
		uintptr(unsafe.Pointer(&utf16ptr)),
	)

	if com.HRESULT(r) != com.HRESULT(windows.S_OK) {
		err = fmt.Errorf("IMMDevice::GetId failed with code: 0x%08X", com.HRESULT(r))
		return
	}

	id = windows.UTF16PtrToString(utf16ptr)

	com.CoTaskMemFree(unsafe.Pointer(utf16ptr))

	return
}

func (device *IMMDevice) GetState() (state uint32, err error) {
	r, _, _ := syscall.SyscallN(device.vtbl.GetState, uintptr(unsafe.Pointer(device)),
		uintptr(unsafe.Pointer(&state)),
	)

	if com.HRESULT(r) != com.HRESULT(windows.S_OK) {
		err = fmt.Errorf("IMMDevice::GetState failed with code: 0x%08X", com.HRESULT(r))
		return
	}

	return
}
