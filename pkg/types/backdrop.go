package types

import (
	"reflect"

	"github.com/dodo-cli/dodo-core/pkg/decoder"
)

func NewBackdrop() decoder.Producer {
	return func() (interface{}, decoder.Decoding) {
		target := &Backdrop{Build: &BuildInfo{}}
		return &target, DecodeBackdrop(&target)
	}
}

func DecodeBackdrop(target interface{}) decoder.Decoding {
	// TODO: wtf this cast
	backdrop := *(target.(**Backdrop))

	return decoder.Kinds(map[reflect.Kind]decoder.Decoding{
		reflect.Map: decoder.Keys(map[string]decoder.Decoding{
			"alias":   decoder.Slice(decoder.NewString(), &backdrop.Aliases),
			"aliases": decoder.Slice(decoder.NewString(), &backdrop.Aliases),
			"image":   DecodeBuildInfo(&backdrop.Build),
			"build":   DecodeBuildInfo(&backdrop.Build),
		}),
	})
}
