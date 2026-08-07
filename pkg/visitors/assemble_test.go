// Copyright 2019 the LinuxBoot Authors. All rights reserved
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package visitors

import (
	"fmt"
	"testing"

	"github.com/linuxboot/fiano/pkg/guid"
	"github.com/linuxboot/fiano/pkg/uefi"
)

var (
	ZeroGUID = guid.MustParse("00000000-0000-0000-0000-000000000000")
)

func TestBadDepex(t *testing.T) {
	var tests = []struct {
		name string
		op   uefi.DepExOp
		err  string
	}{
		{"badOpCode", uefi.DepExOp{OpCode: "BLAH", GUID: nil},
			"unable to map depex opcode string to opcode, string was: BLAH"},
		{"pushNoGUID", uefi.DepExOp{OpCode: "PUSH", GUID: nil},
			"depex opcode PUSH must not have nil guid"},
		{"trueWithGUID", uefi.DepExOp{OpCode: "TRUE", GUID: ZeroGUID},
			fmt.Sprintf("depex opcode TRUE must not have a guid! got %v", *ZeroGUID)},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := &uefi.Section{}
			s.SetType(uefi.SectionTypeDXEDepEx)
			s.DepEx = []uefi.DepExOp{test.op}
			a := &Assemble{}
			err := a.Run(s)
			if err == nil {
				t.Fatalf("Expected error: %v, got nil!", test.err)
			}
			if errStr := err.Error(); test.err != errStr {
				t.Errorf("Expected error: %v, got %v instead", test.err, errStr)
			}
		})
	}
}

func TestAssembleRegeneratesFirmwareVolumePadFiles(t *testing.T) {
	const (
		firmwareVolumeSize = 0x1000
		originalFileSize   = 0x40
		grownFileSize      = 0xa0
		alignedFileSize    = 0x40
		expectedPadSize    = 0x60
	)

	originalErasePolarity := uefi.Attributes.ErasePolarity
	uefi.Attributes.ErasePolarity = 0xff
	t.Cleanup(func() {
		uefi.Attributes.ErasePolarity = originalErasePolarity
	})

	fv, err := createEmptyFirmwareVolume(
		0,
		firmwareVolumeSize,
		nil,
	)
	if err != nil {
		t.Fatalf(
			"createEmptyFirmwareVolume() returned an error: %v",
			err,
		)
	}

	first := newTestRawFile(
		t,
		"11111111-1111-1111-1111-111111111111",
		originalFileSize,
		false,
	)
	second := newTestRawFile(
		t,
		"22222222-2222-2222-2222-222222222222",
		alignedFileSize,
		true,
	)

	fv.Files = []*uefi.File{
		first,
		second,
	}

	if err := (&Assemble{}).Run(fv); err != nil {
		t.Fatalf(
			"initial Assemble.Run() returned an error: %v",
			err,
		)
	}

	parsed, err := uefi.NewFirmwareVolume(
		fv.Buf(),
		0,
		false,
	)
	if err != nil {
		t.Fatalf(
			"parse initially assembled firmware volume: %v",
			err,
		)
	}

	if len(parsed.Files) != 3 {
		t.Fatalf(
			"initially assembled file count = %d, want 3",
			len(parsed.Files),
		)
	}

	pad := parsed.Files[1]

	if pad.Header.Type != uefi.FVFileTypePad {
		t.Fatalf(
			"initial middle file type = %s, want %s",
			pad.Header.Type.String(),
			uefi.FVFileTypePad.String(),
		)
	}

	if pad.Header.ExtendedSize != expectedPadSize {
		t.Fatalf(
			"initial PAD file size = %#x, want %#x",
			pad.Header.ExtendedSize,
			expectedPadSize,
		)
	}

	resizeTestRawFile(
		t,
		parsed.Files[0],
		grownFileSize,
	)

	if err := (&Assemble{}).Run(parsed); err != nil {
		t.Fatalf(
			"rebuilt Assemble.Run() returned an error: %v",
			err,
		)
	}

	rebuilt, err := uefi.NewFirmwareVolume(
		parsed.Buf(),
		0,
		false,
	)
	if err != nil {
		t.Fatalf(
			"parse rebuilt firmware volume: %v",
			err,
		)
	}

	if len(rebuilt.Files) != 2 {
		t.Fatalf(
			"rebuilt file count = %d, want 2",
			len(rebuilt.Files),
		)
	}

	for index, file := range rebuilt.Files {
		if file.Header.Type == uefi.FVFileTypePad {
			t.Fatalf(
				"rebuilt file %d is an obsolete PAD file of %#x bytes",
				index,
				file.Header.ExtendedSize,
			)
		}
	}

	if rebuilt.Files[0].Header.ExtendedSize != grownFileSize {
		t.Fatalf(
			"rebuilt first file size = %#x, want %#x",
			rebuilt.Files[0].Header.ExtendedSize,
			grownFileSize,
		)
	}

	if rebuilt.Files[1].Header.ExtendedSize != alignedFileSize {
		t.Fatalf(
			"rebuilt second file size = %#x, want %#x",
			rebuilt.Files[1].Header.ExtendedSize,
			alignedFileSize,
		)
	}
}

func TestAssemblePreservesCreatedFirmwareVolumePadFile(t *testing.T) {
	const (
		firmwareVolumeSize = 0x1000
		padSize            = 0x100
		fileSize           = 0x40
	)

	originalErasePolarity := uefi.Attributes.ErasePolarity
	uefi.Attributes.ErasePolarity = 0xff
	t.Cleanup(func() {
		uefi.Attributes.ErasePolarity = originalErasePolarity
	})

	fv, err := createEmptyFirmwareVolume(
		0,
		firmwareVolumeSize,
		nil,
	)
	if err != nil {
		t.Fatalf(
			"createEmptyFirmwareVolume() returned an error: %v",
			err,
		)
	}

	pad, err := uefi.CreatePadFile(padSize)
	if err != nil {
		t.Fatalf(
			"CreatePadFile() returned an error: %v",
			err,
		)
	}

	file := newTestRawFile(
		t,
		"33333333-3333-3333-3333-333333333333",
		fileSize,
		false,
	)

	fv.Files = []*uefi.File{
		pad,
		file,
	}

	if err := (&Assemble{}).Run(fv); err != nil {
		t.Fatalf(
			"Assemble.Run() returned an error: %v",
			err,
		)
	}

	parsed, err := uefi.NewFirmwareVolume(
		fv.Buf(),
		0,
		false,
	)
	if err != nil {
		t.Fatalf(
			"parse assembled firmware volume: %v",
			err,
		)
	}

	if len(parsed.Files) != 2 {
		t.Fatalf(
			"assembled file count = %d, want 2",
			len(parsed.Files),
		)
	}

	if parsed.Files[0].Header.Type != uefi.FVFileTypePad {
		t.Fatalf(
			"first assembled file type = %s, want %s",
			parsed.Files[0].Header.Type.String(),
			uefi.FVFileTypePad.String(),
		)
	}

	if parsed.Files[0].Header.ExtendedSize != padSize {
		t.Fatalf(
			"assembled PAD file size = %#x, want %#x",
			parsed.Files[0].Header.ExtendedSize,
			padSize,
		)
	}
}

func newTestRawFile(
	t *testing.T,
	fileGUID string,
	size uint64,
	align128 bool,
) *uefi.File {
	t.Helper()

	file := &uefi.File{}
	file.Header.GUID = *guid.MustParse(fileGUID)
	file.Header.Type = uefi.FVFileTypeRaw
	file.Type = file.Header.Type.String()

	if align128 {
		file.Header.Attributes = 0x10
	}

	resizeTestRawFile(
		t,
		file,
		size,
	)

	file.Modified = false

	return file
}

func resizeTestRawFile(
	t *testing.T,
	file *uefi.File,
	size uint64,
) {
	t.Helper()

	if file == nil {
		t.Fatal("raw file is nil")
	}

	if size < uefi.FileHeaderMinLength {
		t.Fatalf(
			"raw file size = %#x, want at least %#x",
			size,
			uefi.FileHeaderMinLength,
		)
	}

	file.SetSize(
		size,
		true,
	)
	file.Header.SetState(
		uefi.FileStateValid,
	)

	dataLength := size - file.HeaderLen()
	data := make(
		[]byte,
		dataLength,
	)

	for index := range data {
		data[index] = byte(index)
	}

	if err := file.ChecksumAndAssemble(data); err != nil {
		t.Fatalf(
			"ChecksumAndAssemble() returned an error: %v",
			err,
		)
	}

	file.Modified = true
}
