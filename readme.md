<h1 align="center">Fiano<br />
<div align="center">
<a href="https://github.com/qemus/fiano"><img src="https://raw.githubusercontent.com/qemus/fiano/master/.github/logo.svg" title="Logo" style="max-width:100%;" width="128" /></a>
</div>
<div align="center">
  
[![Build]][build_url]
[![Version]][release_url]
[![Size]][release_url]

</div></h1>

Prebuilt multi-platform binaries of [Fiano](https://github.com/linuxboot/fiano), a tool for modifying UEFI firmware.

## UTK: Generic UEFI tool kit meant to handle rom images

Example usage:

```
# For a comprehensive list of commands
utk -h

# Display the image in a compact table form:
utk winterfell.rom table

# Summarize everything in JSON:
utk winterfell.rom json

# List information about a single file in JSON (using regex):
utk winterfell.rom find Shell

# Dump an EFI file to an ffs
utk winterfell.rom dump DxeCore dxecore.ffs

# Insert an EFI file into an FV near another Dxe
utk winterfell.rom insert_before Shell dxecore.ffs save inserted.rom
utk winterfell.rom insert_after Shell dxecore.ffs save inserted.rom

# Insert an EFI file into an FV at the front or the end
# "Shell" is just a means of specifying the FV that contains Shell
utk winterfell.rom insert_front Shell dxecore.ffs save inserted.rom
utk winterfell.rom insert_end Shell dxecore.ffs save inserted.rom

# Remove a file and pad the firmware volume to maintain offsets for the following files
utk winterfell.rom remove_pad Shell save removed.rom

# Remove two files by their GUID without padding and replace shell with Linux:
utk winterfell.rom \
  remove 12345678-9abc-def0-1234-567890abcdef \
  remove 23830293-3029-3823-0922-328328330939 \
  replace_pe32 Shell linux.efi \
  save winterfell2.rom

# Extract everything into a directory:
utk winterfell.rom extract winterfell/

# Re-assemble the directory into an image:
utk winterfell/ save winterfell2.rom
```

## Stars 🌟
[![Stars](https://starchart.cc/qemus/fiano.svg?variant=adaptive)](https://starchart.cc/qemus/fiano)

[build_url]: https://github.com/qemus/fiano/
[release_url]: https://github.com/qemus/fiano/releases

[Build]: https://github.com/qemus/fiano/actions/workflows/build.yml/badge.svg
[Size]: https://img.shields.io/badge/size-4.14_MB-steelblue?style=flat&color=066da5
[Version]: https://img.shields.io/github/v/tag/qemus/fiano?label=version&sort=semver&color=066da5
