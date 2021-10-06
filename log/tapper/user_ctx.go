package tapper

import (
	"context"

	"github.com/gin-gonic/gin"
)

type UserContext struct {
	LogUnit *LogUnits
	GinCtx  context.Context
	Error   error
}

type options struct {
	UrlPath string
	GinCtx  context.Context
	Error   error
}

type Option func(*options)

func WithGinCtx(ctx context.Context) Option {
	return func(opt *options) {
		opt.GinCtx = ctx
	}
}

func WithUrlPath(urlPath string) Option {
	return func(opt *options) {
		opt.UrlPath = urlPath
	}
}

func WithError(err error) Option {
	return func(opt *options) {
		opt.Error = err
	}
}

func NewUserContext(opts ...Option) *UserContext {
	logUnit := &LogUnits{}
	o := &options{}
	for _, opt := range opts {
		opt(o)
	}
	logUnit.SetBeginTime()
	logUnit.SetCmd(o.UrlPath)
	return &UserContext{LogUnit: logUnit, GinCtx: o.GinCtx}
}

func (this *UserContext) SetErr(err error) {
	this.LogUnit.AddLogUnit("err_info", err.Error())
}

func CtxTransfer(ctx context.Context, key string) *UserContext {
	ginCtx, ok := ctx.(*gin.Context)
	if !ok {
		return NewUserContext(WithGinCtx(ctx))
	}
	l, ok := ginCtx.Get(key)
	if !ok {
		return NewUserContext(WithGinCtx(ctx))
	}
	userCtx, ok := l.(*UserContext)
	if !ok {
		return NewUserContext(WithGinCtx(ctx))
	}
	return userCtx
}

// gin.Context 属于对象池，因此在异步执行时，需要拷贝一个再使用.注:在多个goroutine中，应使用线程安全的AddLogUnit方法
func (this *UserContext) Clone() *UserContext {
	if this == nil {
		return nil
	}
	q := &UserContext{}
	if ginCtx, ok := this.GinCtx.(*gin.Context); ok {
		q.GinCtx = ginCtx.Copy()
	} else {
		q.GinCtx = &gin.Context{}
	}
	q.LogUnit = this.LogUnit

	return q
}
