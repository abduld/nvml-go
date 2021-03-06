package nvml

// #cgo LDFLAGS: -lnvidia-ml
// #cgo CFLAGS: -I/home/abduld/usr/nvml/include -L/home/abduld/usr/nvml/lib64
// #include <stdio.h>
// #include <stdlib.h>
// #include <nvml.h>
import "C"

import (
	"fmt"
	"strconv"
	"unsafe"
)

type ComputeMode C.nvmlComputeMode_t
type Feature uint
type ECCBitType uint
type ECCCounterType uint
type ClockType uint
type DriverModel uint
type PState C.nvmlPstates_t
type InformObject uint
type Result struct {
	code C.nvmlReturn_t
}

func (c Feature) String() string {
	if c == 0 {
		return "Disabled"
	} else {
		return "Enabled"
	}
}

func (c ComputeMode) String() string {
	switch c {
	case 0:
		return "Default"
	case 1:
		return "ExclusiveThread"
	case 2:
		return "Prohibited"
	case 3:
		return "ExclusiveProcess"
	}
	return fmt.Sprint("UnknownComputeMode %d", c)
}

func (e ECCBitType) String() string {
	if e == 0 {
		return "SingleBitECC"
	} else {
		return "DoubleBitECC"
	}
}

func (e ECCCounterType) String() string {
	if e == 0 {
		return "VolatileECC"
	} else {
		return "AggregateECC"
	}
}

func (c ClockType) String() string {
	switch c {
	case 0:
		return "Graphics"
	case 1:
		return "SM"
	case 2:
		return "Memory"
	}
	return fmt.Sprint("UnknownClockType %d", c)
}

func (d DriverModel) String() string {
	if d == 0 {
		return "WDDM"
	} else {
		return "WDM"
	}
}

func (p PState) String() string {
	if p >= 0 && p < 16 {
		return strconv.Itoa(int(p))
	} else if p == 32 {
		return "Unknown"
	}
	return fmt.Sprint("UnknownPState %d", p)
}

func (i InformObject) String() string {
	switch i {
	case 0:
		return "OEM"
	case 1:
		return "ECC"
	case 2:
		return "Power"
	}
	return fmt.Sprint("UnknownInformObject %d", i)
}

func (r Result) String() string {
	switch r.code {
	case 0:
		return "Success"
	case 1:
		return "Uninitialized"
	case 2:
		return "InvalidArgument"
	case 3:
		return "NotSupported"
	case 4:
		return "NoPermission"
	case 5:
		return "AlreadyInitialized"
	case 6:
		return "NotFound"
	case 7:
		return "InsufficientSize"
	case 8:
		return "InsufficientPower"
	case 9:
		return "DriverNotLoaded"
	case 10:
		return "Timeout"
	case 99:
		return "Unknown"
	}
	return fmt.Sprint("UnknownError %d", r)
}

func (r Result) Error() string {
	return r.String()
}

func (r Result) SuccessQ() bool {
	if r.code == 0 {
		return true
	} else {
		return false
	}
}

func NewResult(r C.nvmlReturn_t) error {
	if r == 0 {
		return nil
	} else {
		return &Result{r}
	}
}

func Init() error {
	r := C.nvmlInit()
	return NewResult(r)
}

func Shutdown() error {
	r := C.nvmlShutdown()
	return NewResult(r)
}

func ErrorString(r Result) string {
	s := C.nvmlErrorString(r.code)
	return C.GoString(s)
}

func DeviceCount() (int, error) {
	var count C.uint = 0
	r := NewResult(C.nvmlDeviceGetCount(&count))
	return int(count), r
}

type DeviceHandle struct {
	handle C.nvmlDevice_t
}

func DeviceGetHandleByIndex(idx int) (DeviceHandle, error) {
	var device C.nvmlDevice_t
	r := NewResult(C.nvmlDeviceGetHandleByIndex(C.uint(idx), &device))
	return DeviceHandle{device}, r
}

//compute mode

func DeviceComputeMode(dh DeviceHandle) (ComputeMode, error) {
	var mode C.nvmlComputeMode_t
	r := NewResult(C.nvmlDeviceGetComputeMode(dh.handle, &mode))
	return ComputeMode(mode), r
}

//device name

const STRING_BUFFER_SIZE = 100

func makeStringBuffer(sz int) *C.char {
	b := make([]byte, sz)
	return C.CString(string(b))
}

func DeviceName(dh DeviceHandle) (string, error) {
	var name *C.char = makeStringBuffer(STRING_BUFFER_SIZE)
	defer C.free(unsafe.Pointer(name))
	r := NewResult(C.nvmlDeviceGetName(dh.handle, name, C.uint(STRING_BUFFER_SIZE)))
	return C.GoStringN(name, STRING_BUFFER_SIZE), r
}

type MemoryInformation struct {
	Used  uint64 `json:"used"`
	Free  uint64 `json:"free"`
	Total uint64 `json:"total"`
}

func DeviceMemoryInformation(dh DeviceHandle) (MemoryInformation, error) {
	var temp C.nvmlMemory_t
	r := NewResult(C.nvmlDeviceGetMemoryInfo(dh.handle, &temp))
	if r == nil {
		res := MemoryInformation{
			Used:  uint64(temp.used),
			Free:  uint64(temp.free),
			Total: uint64(temp.total),
		}
		return res, nil
	}
	return MemoryInformation{}, r
}

type PCIInformation struct {
	BusId       string `json:"bus_id"`
	Domain      uint   `json:"domain"`
	Bus         uint   `json:"bus"`
	Device      uint   `json:"device"`
	DeviceId    uint   `json:"device_id"`
	SubSystemId uint   `json:"subsystem_id"`
}

func DevicePCIInformation(dh DeviceHandle) (PCIInformation, error) {
	var temp C.nvmlPciInfo_t
	r := NewResult(C.nvmlDeviceGetPciInfo(dh.handle, &temp))
	if r == nil {
		res := PCIInformation{
			BusId: string(C.GoBytes(unsafe.Pointer(&temp.busId),
				C.NVML_DEVICE_PCI_BUS_ID_BUFFER_SIZE)),
			Domain:      uint(temp.domain),
			Bus:         uint(temp.bus),
			Device:      uint(temp.device),
			DeviceId:    uint(temp.pciDeviceId),
			SubSystemId: uint(temp.pciSubSystemId),
		}
		return res, nil
	}
	return PCIInformation{}, r
}

func DeviceTemperature(dh DeviceHandle) (uint, error) {
	var temp C.uint
	r := NewResult(C.nvmlDeviceGetTemperature(dh.handle, C.nvmlTemperatureSensors_t(0), &temp))
	return uint(temp), r
}

func DevicePerformanceState(dh DeviceHandle) (PState, error) {
	var pstate C.nvmlPstates_t
	r := NewResult(C.nvmlDeviceGetPerformanceState(dh.handle, &pstate))
	return PState(pstate), r
}

func DeviceFanSpeed(dh DeviceHandle) (uint, error) {
	var speed C.uint
	r := NewResult(C.nvmlDeviceGetFanSpeed(dh.handle, &speed))
	return uint(speed), r
}

func main() {
	Init()
}
