package com

import (
	"fmt"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

type HRESULT = uint32

var (
	ole32                = windows.NewLazySystemDLL("ole32.dll")
	procCoCreateInstance = ole32.NewProc("CoCreateInstance")
	procCoInitializeEx   = ole32.NewProc("CoInitializeEx")
	procCoUninitialize   = ole32.NewProc("CoUninitialize")
	procCoTaskMemFree    = ole32.NewProc("CoTaskMemFree")
)

func CoInitializeEx(reserved uintptr, coInitFlag uint32) (err error) {
	r, _, _ := syscall.SyscallN(procCoInitializeEx.Addr(),
		reserved,
		uintptr(coInitFlag),
	)

	if HRESULT(r) != HRESULT(windows.S_OK) {
		err = fmt.Errorf("com::CoInitializeEx failed with code: 0x%08X", HRESULT(r))
		return
	}

	return
}

func CoTaskMemFree(address unsafe.Pointer) {
	syscall.SyscallN(procCoTaskMemFree.Addr(), uintptr(address))
}

func CoUninitialize() {
	syscall.SyscallN(procCoUninitialize.Addr())
}

func CoCreateInstance(
	clsid *windows.GUID,
	unkOuter *IUnknown,
	clsContext uint32,
	iid *windows.GUID,
) (v unsafe.Pointer, err error) {
	r, _, _ := syscall.SyscallN(procCoCreateInstance.Addr(),
		uintptr(unsafe.Pointer(clsid)),
		uintptr(unsafe.Pointer(unkOuter)),
		uintptr(clsContext),
		uintptr(unsafe.Pointer(iid)),
		uintptr(unsafe.Pointer(&v)),
	)

	if HRESULT(r) != HRESULT(windows.S_OK) {
		err = fmt.Errorf("com::CoCreateInstance failed with code: 0x%08X", HRESULT(r))
		return
	}

	return
}
