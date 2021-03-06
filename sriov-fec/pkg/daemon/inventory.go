// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2020 Intel Corporation

package daemon

import (
	"errors"

	"github.com/go-logr/logr"
	"github.com/intel/sriov-network-device-plugin/pkg/utils"
	"github.com/jaypipes/ghw"
	sriovv1 "github.com/open-ness/openshift-operator/sriov-fec/api/v1"
)

const (
	acceleratorClass    = "12"
	acceleratorSubclass = "00"
	vendorID            = "8086"
)

type deviceInfo struct {
	VFDeviceID string
	DeviceName string
}

var (
	deviceIDWhitelist = map[string]deviceInfo{
		"0d8f": {"0d90", "FPGA_5GNR"},
		"5052": {"5050", "FPGA_LTE"},
		"0b32": {"", ""}, // Factory dummy
	}
)

func GetSriovInventory(log logr.Logger) (*sriovv1.NodeInventory, error) {
	pci, err := ghw.PCI()
	if err != nil {
		log.Error(err, "failed to get PCI info")
		return nil, err
	}

	devices := pci.ListDevices()
	if len(devices) == 0 {
		log.Info("got 0 pci devices")
		err := errors.New("pci.ListDevices() returned 0 devices")
		return nil, err
	}

	accelerators := &sriovv1.NodeInventory{
		SriovAccelerators: []sriovv1.SriovAccelerator{},
	}

	for _, device := range devices {
		if !(device.Vendor.ID == vendorID &&
			device.Class.ID == acceleratorClass &&
			device.Subclass.ID == acceleratorSubclass) {
			continue
		}

		if _, ok := deviceIDWhitelist[device.Product.ID]; !ok {
			continue
		}

		if !utils.IsSriovPF(device.Address) {
			continue
		}

		driver, err := utils.GetDriverName(device.Address)
		if err != nil {
			log.Info("unable to get driver for device", "pci", device.Address, "reason", err.Error())
			driver = ""
		}

		acc := sriovv1.SriovAccelerator{
			VendorID:   device.Vendor.ID,
			DeviceID:   device.Product.ID,
			PCIAddress: device.Address,
			Driver:     driver,
			MaxVFs:     utils.GetSriovVFcapacity(device.Address),
			VFs:        []sriovv1.VF{},
		}

		vfs, err := utils.GetVFList(device.Address)
		if err != nil {
			log.Error(err, "failed to get list of VFs for device", "pci", device.Address)
		}

		for _, vf := range vfs {
			vfInfo := sriovv1.VF{
				PCIAddress: vf,
			}

			driver, err := utils.GetDriverName(vf)
			if err != nil {
				log.Info("failed to get driver name for VF", "pci", vf, "pf", device.Address, "reason", err.Error())
			} else {
				vfInfo.Driver = driver
			}

			if vfDeviceInfo := pci.GetDevice(vf); vfDeviceInfo == nil {
				log.Info("failed to get device info for vf", "pci", vf)
			} else {
				vfInfo.DeviceID = vfDeviceInfo.Product.ID
			}

			acc.VFs = append(acc.VFs, vfInfo)
		}

		accelerators.SriovAccelerators = append(accelerators.SriovAccelerators, acc)
	}

	return accelerators, nil
}
