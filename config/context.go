package config

import "context"

type ctxKey int

const (
	ctxKeyOutputDir ctxKey = iota + 1
	ctxKeyLockfile
)

func WithOutputDir(ctx context.Context, outputDir string) context.Context {
	return context.WithValue(ctx, ctxKeyOutputDir, outputDir)
}

func CtxOutputDir(ctx context.Context) (string, bool) {
	d, ok := ctx.Value(ctxKeyOutputDir).(string)
	return d, ok
}

func WithLockfile(ctx context.Context, lockfile *Lockfile) context.Context {
	return context.WithValue(ctx, ctxKeyLockfile, lockfile)
}

func CtxLockfile(ctx context.Context) (*Lockfile, bool) {
	l, ok := ctx.Value(ctxKeyLockfile).(*Lockfile)
	return l, ok
}
