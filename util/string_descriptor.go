package util

import "context"

type stringDescriptor struct{ value string }

func (d stringDescriptor) Description(ctx context.Context) string {
	return d.MarkdownDescription(ctx)
}

func (d stringDescriptor) MarkdownDescription(_ context.Context) string {
	return d.value
}
