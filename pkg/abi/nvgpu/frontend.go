// Copyright 2023 The gVisor Authors.
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

package nvgpu

import (
	"gvisor.dev/gvisor/pkg/marshal"
)

// NV_IOCTL_MAGIC is the "canonical" IOC_TYPE for frontend ioctls.
// The driver ignores IOC_TYPE, allowing any value to be passed.
const NV_IOCTL_MAGIC = uint32('F')

// Frontend ioctl numbers.
// Note that these are only the IOC_NR part of the ioctl command.
const (
	// From kernel-open/common/inc/nv-ioctl-numbers.h:
	NV_IOCTL_BASE            = 200
	NV_ESC_CARD_INFO         = NV_IOCTL_BASE + 0
	NV_ESC_REGISTER_FD       = NV_IOCTL_BASE + 1
	NV_ESC_ALLOC_OS_EVENT    = NV_IOCTL_BASE + 6
	NV_ESC_FREE_OS_EVENT     = NV_IOCTL_BASE + 7
	NV_ESC_CHECK_VERSION_STR = NV_IOCTL_BASE + 10
	NV_ESC_SYS_PARAMS        = NV_IOCTL_BASE + 14

	// From kernel-open/common/inc/nv-ioctl-numa.h:
	NV_ESC_NUMA_INFO = NV_IOCTL_BASE + 15

	// From src/nvidia/arch/nvalloc/unix/include/nv_escape.h:
	NV_ESC_RM_ALLOC_MEMORY               = 0x27
	NV_ESC_RM_FREE                       = 0x29
	NV_ESC_RM_CONTROL                    = 0x2a
	NV_ESC_RM_ALLOC                      = 0x2b
	NV_ESC_RM_DUP_OBJECT                 = 0x34
	NV_ESC_RM_SHARE                      = 0x35
	NV_ESC_RM_VID_HEAP_CONTROL           = 0x4a
	NV_ESC_RM_MAP_MEMORY                 = 0x4e
	NV_ESC_RM_UNMAP_MEMORY               = 0x4f
	NV_ESC_RM_UPDATE_DEVICE_MAPPING_INFO = 0x5e
)

// Frontend ioctl parameter structs, from src/common/sdk/nvidia/inc/nvos.h or
// kernel-open/common/inc/nv-ioctl.h.

// IoctlRegisterFD is nv_ioctl_register_fd_t, the parameter type for
// NV_ESC_REGISTER_FD.
//
// +marshal
type IoctlRegisterFD struct {
	CtlFD int32
}

// IoctlAllocOSEvent is nv_ioctl_alloc_os_event_t, the parameter type for
// NV_ESC_ALLOC_OS_EVENT.
//
// +marshal
type IoctlAllocOSEvent struct {
	HClient Handle
	HDevice Handle
	FD      uint32
	Status  uint32
}

// IoctlFreeOSEvent is nv_ioctl_free_os_event_t, the parameter type for
// NV_ESC_FREE_OS_EVENT.
//
// +marshal
type IoctlFreeOSEvent struct {
	HClient Handle
	HDevice Handle
	FD      uint32
	Status  uint32
}

// RMAPIVersion is nv_rm_api_version_t, the parameter type for
// NV_ESC_CHECK_VERSION_STR.
//
// +marshal
type RMAPIVersion struct {
	Cmd           uint32
	Reply         uint32
	VersionString [64]byte
}

// IoctlSysParams is nv_ioctl_sys_params_t, the parameter type for
// NV_ESC_SYS_PARAMS.
//
// +marshal
type IoctlSysParams struct {
	MemblockSize uint64
}

// IoctlNVOS02ParametersWithFD is nv_ioctl_nvos2_parameters_with_fd, the
// parameter type for NV_ESC_RM_ALLOC_MEMORY.
//
// +marshal
type IoctlNVOS02ParametersWithFD struct {
	Params NVOS02Parameters
	FD     int32
	Pad0   [4]byte
}

// +marshal
type NVOS02Parameters struct {
	HRoot         Handle
	HObjectParent Handle
	HObjectNew    Handle
	HClass        uint32
	Flags         uint32
	Pad0          [4]byte
	PMemory       P64 // address of application mapping, without indirection
	Limit         uint64
	Status        uint32
	Pad1          [4]byte
}

// NVOS00Parameters is NVOS00_PARAMETERS, the parameter type for
// NV_ESC_RM_FREE.
//
// +marshal
type NVOS00Parameters struct {
	HRoot         Handle
	HObjectParent Handle
	HObjectOld    Handle
	Status        uint32
}

// RmAllocParamType should be implemented by all possible parameter types for
// NV_ESC_RM_ALLOC.
type RmAllocParamType interface {
	GetPAllocParms() P64
	GetPRightsRequested() P64
	SetPAllocParms(p P64)
	SetPRightsRequested(p P64)
	FromOS64V535(other NVOS64ParametersV535)
	ToOS64V535() NVOS64ParametersV535
	GetPointer() uintptr
	marshal.Marshallable
}

// GetRmAllocParamObj returns the appropriate implementation of
// RmAllocParamType based on passed parameters.
func GetRmAllocParamObj(isNVOS64 bool, isV535 bool) RmAllocParamType {
	if isNVOS64 {
		if isV535 {
			return &NVOS64ParametersV535{}
		}
		return &NVOS64Parameters{}
	}
	if isV535 {
		return &NVOS21ParametersV535{}
	}
	return &NVOS21Parameters{}
}

// NVOS21Parameters is NVOS21_PARAMETERS, one possible parameter type for
// NV_ESC_RM_ALLOC.
//
// +marshal
type NVOS21Parameters struct {
	HRoot         Handle
	HObjectParent Handle
	HObjectNew    Handle
	HClass        uint32
	PAllocParms   P64
	Status        uint32
	Pad0          [4]byte
}

// GetPAllocParms implements RmAllocParamType.GetPAllocParms.
func (n *NVOS21Parameters) GetPAllocParms() P64 {
	return n.PAllocParms
}

// GetPRightsRequested implements RmAllocParamType.GetPRightsRequested.
func (n *NVOS21Parameters) GetPRightsRequested() P64 {
	return 0
}

// SetPAllocParms implements RmAllocParamType.SetPAllocParms.
func (n *NVOS21Parameters) SetPAllocParms(p P64) { n.PAllocParms = p }

// SetPRightsRequested implements RmAllocParamType.SetPRightsRequested.
func (n *NVOS21Parameters) SetPRightsRequested(p P64) {
	panic("impossible")
}

// FromOS64V535 implements RmAllocParamType.FromOS64V535.
func (n *NVOS21Parameters) FromOS64V535(other NVOS64ParametersV535) {
	n.HRoot = other.HRoot
	n.HObjectParent = other.HObjectParent
	n.HObjectNew = other.HObjectNew
	n.HClass = other.HClass
	n.PAllocParms = other.PAllocParms
	n.Status = other.Status
}

// ToOS64V535 implements RmAllocParamType.ToOS64V535.
func (n *NVOS21Parameters) ToOS64V535() NVOS64ParametersV535 {
	return NVOS64ParametersV535{
		HRoot:         n.HRoot,
		HObjectParent: n.HObjectParent,
		HObjectNew:    n.HObjectNew,
		HClass:        n.HClass,
		PAllocParms:   n.PAllocParms,
		Status:        n.Status,
	}
}

// NVOS21ParametersV535 is the updated version of NVOS21Parameters starting
// from 535.43.02.
//
// +marshal
type NVOS21ParametersV535 struct {
	HRoot         Handle
	HObjectParent Handle
	HObjectNew    Handle
	HClass        uint32
	PAllocParms   P64
	ParamsSize    uint32
	Status        uint32
}

// GetPAllocParms implements RmAllocParamType.GetPAllocParms.
func (n *NVOS21ParametersV535) GetPAllocParms() P64 {
	return n.PAllocParms
}

// GetPRightsRequested implements RmAllocParamType.GetPRightsRequested.
func (n *NVOS21ParametersV535) GetPRightsRequested() P64 {
	return 0
}

// SetPAllocParms implements RmAllocParamType.SetPAllocParms.
func (n *NVOS21ParametersV535) SetPAllocParms(p P64) { n.PAllocParms = p }

// SetPRightsRequested implements RmAllocParamType.SetPRightsRequested.
func (n *NVOS21ParametersV535) SetPRightsRequested(p P64) {
	panic("impossible")
}

// FromOS64V535 implements RmAllocParamType.FromOS64V535.
func (n *NVOS21ParametersV535) FromOS64V535(other NVOS64ParametersV535) {
	n.HRoot = other.HRoot
	n.HObjectParent = other.HObjectParent
	n.HObjectNew = other.HObjectNew
	n.HClass = other.HClass
	n.PAllocParms = other.PAllocParms
	n.ParamsSize = other.ParamsSize
	n.Status = other.Status
}

// ToOS64V535 implements RmAllocParamType.ToOS64V535.
func (n *NVOS21ParametersV535) ToOS64V535() NVOS64ParametersV535 {
	return NVOS64ParametersV535{
		HRoot:         n.HRoot,
		HObjectParent: n.HObjectParent,
		HObjectNew:    n.HObjectNew,
		HClass:        n.HClass,
		PAllocParms:   n.PAllocParms,
		ParamsSize:    n.ParamsSize,
		Status:        n.Status,
	}
}

// NVOS55Parameters is NVOS55_PARAMETERS, the parameter type for
// NV_ESC_RM_DUP_OBJECT.
//
// +marshal
type NVOS55Parameters struct {
	HClient    Handle
	HParent    Handle
	HObject    Handle
	HClientSrc Handle
	HObjectSrc Handle
	Flags      uint32
	Status     uint32
}

// NVOS57Parameters is NVOS57_PARAMETERS, the parameter type for
// NV_ESC_RM_SHARE.
//
// +marshal
type NVOS57Parameters struct {
	HClient     Handle
	HObject     Handle
	SharePolicy RS_SHARE_POLICY
	Status      uint32
}

// NVOS32Parameters is NVOS32_PARAMETERS, the parameter type for
// NV_ESC_RM_VID_HEAP_CONTROL.
//
// +marshal
type NVOS32Parameters struct {
	HRoot         Handle
	HObjectParent Handle
	Function      uint32
	HVASpace      Handle
	IVCHeapNumber int16
	Pad           [2]byte
	Status        uint32
	Total         uint64
	Free          uint64
	Data          [144]byte // union
}

// Possible values for NVOS32Parameters.Function:
const (
	NVOS32_FUNCTION_ALLOC_SIZE = 2
)

// NVOS32AllocSize is the type of NVOS32Parameters.Data for
// NVOS32_FUNCTION_ALLOC_SIZE.
type NVOS32AllocSize struct {
	Owner           uint32
	HMemory         Handle
	Type            uint32
	Flags           uint32
	Attr            uint32
	Format          uint32
	ComprCovg       uint32
	ZcullCovg       uint32
	PartitionStride uint32
	Width           uint32
	Height          uint32
	Pad0            [4]byte
	Size            uint64
	Alignment       uint64
	Offset          uint64
	Limit           uint64
	Address         P64
	RangeBegin      uint64
	RangeEnd        uint64
	Attr2           uint32
	CtagOffset      uint32
}

// IoctlNVOS33ParametersWithFD is nv_ioctl_nvos33_parameters_with_fd, the
// parameter type for NV_ESC_RM_MAP_MEMORY, from
// src/nvidia/arch/nvalloc/unix/include/nv-unix-nvos-params-wrappers.h.
//
// +marshal
type IoctlNVOS33ParametersWithFD struct {
	Params NVOS33Parameters
	FD     int32
	Pad0   [4]byte
}

// +marshal
type NVOS33Parameters struct {
	HClient        Handle
	HDevice        Handle
	HMemory        Handle
	Pad0           [4]byte
	Offset         uint64
	Length         uint64
	PLinearAddress P64 // address of application mapping, without indirection
	Status         uint32
	Flags          uint32
}

// NVOS34Parameters is NVOS34_PARAMETERS, the parameter type for
// NV_ESC_RM_UNMAP_MEMORY.
//
// +marshal
type NVOS34Parameters struct {
	HClient        Handle
	HDevice        Handle
	HMemory        Handle
	Pad0           [4]byte
	PLinearAddress P64 // address of application mapping, without indirection
	Status         uint32
	Flags          uint32
}

// NVOS54Parameters is NVOS54_PARAMETERS, the parameter type for
// NV_ESC_RM_CONTROL.
//
// +marshal
type NVOS54Parameters struct {
	HClient    Handle
	HObject    Handle
	Cmd        uint32
	Flags      uint32
	Params     P64
	ParamsSize uint32
	Status     uint32
}

// NVOS56Parameters is NVOS56_PARAMETERS, the parameter type for
// NV_ESC_RM_UPDATE_DEVICE_MAPPING_INFO.
//
// +marshal
type NVOS56Parameters struct {
	HClient        Handle
	HDevice        Handle
	HMemory        Handle
	Pad0           [4]byte
	POldCPUAddress P64
	PNewCPUAddress P64
	Status         uint32
	Pad1           [4]byte
}

// NVOS64Parameters is NVOS64_PARAMETERS, one possible parameter type for
// NV_ESC_RM_ALLOC.
//
// +marshal
type NVOS64Parameters struct {
	HRoot            Handle
	HObjectParent    Handle
	HObjectNew       Handle
	HClass           uint32
	PAllocParms      P64
	PRightsRequested P64
	Flags            uint32
	Status           uint32
}

// GetPAllocParms implements RmAllocParamType.GetPAllocParms.
func (n *NVOS64Parameters) GetPAllocParms() P64 {
	return n.PAllocParms
}

// GetPRightsRequested implements RmAllocParamType.GetPRightsRequested.
func (n *NVOS64Parameters) GetPRightsRequested() P64 {
	return n.PRightsRequested
}

// SetPAllocParms implements RmAllocParamType.SetPAllocParms.
func (n *NVOS64Parameters) SetPAllocParms(p P64) { n.PAllocParms = p }

// SetPRightsRequested implements RmAllocParamType.SetPRightsRequested.
func (n *NVOS64Parameters) SetPRightsRequested(p P64) { n.PRightsRequested = p }

// FromOS64V535 implements RmAllocParamType.FromOS64V535.
func (n *NVOS64Parameters) FromOS64V535(other NVOS64ParametersV535) {
	n.HRoot = other.HRoot
	n.HObjectParent = other.HObjectParent
	n.HObjectNew = other.HObjectNew
	n.HClass = other.HClass
	n.PAllocParms = other.PAllocParms
	n.PRightsRequested = other.PRightsRequested
	n.Flags = other.Flags
	n.Status = other.Status
}

// ToOS64V535 implements RmAllocParamType.ToOS64V535.
func (n *NVOS64Parameters) ToOS64V535() NVOS64ParametersV535 {
	return NVOS64ParametersV535{
		HRoot:            n.HRoot,
		HObjectParent:    n.HObjectParent,
		HObjectNew:       n.HObjectNew,
		HClass:           n.HClass,
		PAllocParms:      n.PAllocParms,
		PRightsRequested: n.PRightsRequested,
		Flags:            n.Flags,
		Status:           n.Status,
	}
}

// NVOS64ParametersV535 is the updated version of NVOS64Parameters starting
// from 535.43.02.
//
// +marshal
type NVOS64ParametersV535 struct {
	HRoot            Handle
	HObjectParent    Handle
	HObjectNew       Handle
	HClass           uint32
	PAllocParms      P64
	PRightsRequested P64
	ParamsSize       uint32
	Flags            uint32
	Status           uint32
	_                uint32
}

// GetPAllocParms implements RmAllocParamType.GetPAllocParms.
func (n *NVOS64ParametersV535) GetPAllocParms() P64 {
	return n.PAllocParms
}

// GetPRightsRequested implements RmAllocParamType.GetPRightsRequested.
func (n *NVOS64ParametersV535) GetPRightsRequested() P64 {
	return n.PRightsRequested
}

// SetPAllocParms implements RmAllocParamType.SetPAllocParms.
func (n *NVOS64ParametersV535) SetPAllocParms(p P64) { n.PAllocParms = p }

// SetPRightsRequested implements RmAllocParamType.SetPRightsRequested.
func (n *NVOS64ParametersV535) SetPRightsRequested(p P64) { n.PRightsRequested = p }

// FromOS64V535 implements RmAllocParamType.FromOS64V535.
func (n *NVOS64ParametersV535) FromOS64V535(other NVOS64ParametersV535) { *n = other }

// ToOS64V535 implements RmAllocParamType.ToOS64V535.
func (n *NVOS64ParametersV535) ToOS64V535() NVOS64ParametersV535 { return *n }

// Frontend ioctl parameter struct sizes.
var (
	SizeofIoctlRegisterFD             = uint32((*IoctlRegisterFD)(nil).SizeBytes())
	SizeofIoctlAllocOSEvent           = uint32((*IoctlAllocOSEvent)(nil).SizeBytes())
	SizeofIoctlFreeOSEvent            = uint32((*IoctlFreeOSEvent)(nil).SizeBytes())
	SizeofRMAPIVersion                = uint32((*RMAPIVersion)(nil).SizeBytes())
	SizeofIoctlSysParams              = uint32((*IoctlSysParams)(nil).SizeBytes())
	SizeofIoctlNVOS02ParametersWithFD = uint32((*IoctlNVOS02ParametersWithFD)(nil).SizeBytes())
	SizeofNVOS00Parameters            = uint32((*NVOS00Parameters)(nil).SizeBytes())
	SizeofNVOS21Parameters            = uint32((*NVOS21Parameters)(nil).SizeBytes())
	SizeofNVOS21ParametersV535        = uint32((*NVOS21ParametersV535)(nil).SizeBytes())
	SizeofIoctlNVOS33ParametersWithFD = uint32((*IoctlNVOS33ParametersWithFD)(nil).SizeBytes())
	SizeofNVOS55Parameters            = uint32((*NVOS55Parameters)(nil).SizeBytes())
	SizeofNVOS57Parameters            = uint32((*NVOS57Parameters)(nil).SizeBytes())
	SizeofNVOS32Parameters            = uint32((*NVOS32Parameters)(nil).SizeBytes())
	SizeofNVOS34Parameters            = uint32((*NVOS34Parameters)(nil).SizeBytes())
	SizeofNVOS54Parameters            = uint32((*NVOS54Parameters)(nil).SizeBytes())
	SizeofNVOS56Parameters            = uint32((*NVOS56Parameters)(nil).SizeBytes())
	SizeofNVOS64Parameters            = uint32((*NVOS64Parameters)(nil).SizeBytes())
	SizeofNVOS64ParametersV535        = uint32((*NVOS64ParametersV535)(nil).SizeBytes())
)
