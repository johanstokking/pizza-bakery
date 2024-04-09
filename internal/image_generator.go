package internal

import (
	"context"
	"image"
)

// ImageGenerator generating images from text descriptions.
type ImageGenerator interface {
	Generate(ctx context.Context, description string) (image.Image, error)
}
