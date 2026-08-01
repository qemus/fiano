// Copyright 2026 the LinuxBoot Authors. All rights reserved
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package visitors

import (
	"testing"

	"github.com/linuxboot/fiano/pkg/uefi"
)

func TestSetBufMarksModified(t *testing.T) {
	br := &uefi.BIOSRegion{}
	br.SetBuf(nil)
	if !br.Modified {
		t.Error("BIOSRegion.SetBuf did not mark the region modified")
	}

	fv := &uefi.FirmwareVolume{}
	fv.SetBuf(nil)
	if !fv.Modified {
		t.Error("FirmwareVolume.SetBuf did not mark the volume modified")
	}

	file := &uefi.File{}
	file.SetBuf(nil)
	if !file.Modified {
		t.Error("File.SetBuf did not mark the file modified")
	}

	section := &uefi.Section{}
	section.SetBuf(nil)
	if !section.Modified {
		t.Error("Section.SetBuf did not mark the section modified")
	}

	store := &uefi.NVarStore{}
	store.SetBuf(nil)
	if !store.Modified {
		t.Error("NVarStore.SetBuf did not mark the store modified")
	}

	nvar := &uefi.NVar{}
	nvar.SetBuf(nil)
	if !nvar.Modified {
		t.Error("NVar.SetBuf did not mark the variable modified")
	}
}

func TestPropagateModified(t *testing.T) {
	section := &uefi.Section{Parsed: true, Modified: true}
	file := &uefi.File{
		Parsed:   true,
		Sections: []*uefi.Section{section},
	}
	fv := &uefi.FirmwareVolume{
		Parsed: true,
		Files:  []*uefi.File{file},
	}
	br := &uefi.BIOSRegion{
		Parsed:   true,
		Elements: []*uefi.TypedFirmware{uefi.MakeTyped(fv)},
	}

	if !propagateModified(br) {
		t.Fatal("modified descendant did not mark the root for assembly")
	}

	if !file.Modified || !fv.Modified || !br.Modified {
		t.Fatalf(
			"modified state was not propagated: file=%t fv=%t bios=%t",
			file.Modified,
			fv.Modified,
			br.Modified,
		)
	}
}

func TestFlattenDoesNotMutateFirmware(t *testing.T) {
	inner := &uefi.Section{}
	section := &uefi.Section{
		Encapsulated: []*uefi.TypedFirmware{uefi.MakeTyped(inner)},
	}
	file := &uefi.File{Sections: []*uefi.Section{section}}
	fv := &uefi.FirmwareVolume{Files: []*uefi.File{file}}
	br := &uefi.BIOSRegion{
		Elements: []*uefi.TypedFirmware{uefi.MakeTyped(fv)},
	}

	v := &Flatten{}
	if err := v.Run(br); err != nil {
		t.Fatalf("Flatten.Run failed: %v", err)
	}

	if len(br.Elements) != 1 {
		t.Fatalf("flatten modified BIOSRegion.Elements: got %d", len(br.Elements))
	}
	if len(fv.Files) != 1 {
		t.Fatalf("flatten modified FirmwareVolume.Files: got %d", len(fv.Files))
	}
	if len(file.Sections) != 1 {
		t.Fatalf("flatten modified File.Sections: got %d", len(file.Sections))
	}
	if len(section.Encapsulated) != 1 {
		t.Fatalf("flatten modified Section.Encapsulated: got %d", len(section.Encapsulated))
	}
}
