package internal

import "strings"

func ExtractFileImageTag(ref string) (string, string, string) {
	file, imageOrTag, imageOrTagFound := strings.Cut(ref, ":")

	if !imageOrTagFound {
		return file, "", "latest"
	}

	image, tag, tagFound := strings.Cut(imageOrTag, ":")
	if !tagFound {
		return file, "", image
	} else {
		return file, image, tag
	}
}
