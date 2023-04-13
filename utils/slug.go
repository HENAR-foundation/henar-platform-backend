package utils

import (
	"henar-backend/types"

	"github.com/gosimple/slug"
)

func CreateSlug(Title types.Translations) string {
	var title string
	if Title.En != "" {
		title = Title.En
	} else if Title.Ru != "" {
		title = Title.Ru
	} else if Title.Hy != "" {
		title = Title.Hy
	}

	slugText := slug.Make(title)
	return slugText
}
