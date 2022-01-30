package pipline

import (
	"context"
	"fmt"

	"github.com/zengzhengrong/request"
	"github.com/zengzhengrong/request/opts/client"
	"golang.org/x/sync/errgroup"
)

const (
	piplineCtxValueKey = "values"
)

type (
	In            func(ctx context.Context, client *client.Client) ([]byte, error)
	Ins           []In
	Out           func(ctx context.Context, client *client.Client, In ...[]byte) request.Response
	Parall        bool
	PipLineClient struct{ *client.Client }
)

type PipLineOption interface {
	apply(*PipLine)
}

func (pa Parall) apply(p *PipLine) {
	p.Parall = Parall(pa)
}

func (i Ins) apply(p *PipLine) {
	p.Ins = i
}

func (o Out) apply(p *PipLine) {
	p.Out = o
}

func (c PipLineClient) apply(p *PipLine) {
	p.PipLineClient = PipLineClient(c).Client
}

func WithParall(p bool) PipLineOption {
	return Parall(p)
}

func WithIn(i ...In) PipLineOption {
	return Ins(i)
}

func WithOut(o Out) PipLineOption {
	return Out(o)
}

func WithClient(client *client.Client) PipLineOption {
	return PipLineClient{Client: client}
}

func WithDefaultClient() PipLineOption {
	return PipLineClient{Client: client.NewClient(client.WithDefault())}
}

// pipline 流水线请求, Ins 是先请求的函数返回对应的数据，Out是根据Ins 请求的的数据在组合去请求

type PipLine struct {
	PipLineClient *client.Client
	Parall        Parall
	Ins           Ins
	Out           Out
	Ctx           context.Context
}

func ctxsetfinish(ctx context.Context) context.Context {
	v := ctx.Value(piplineCtxValueKey).(map[string]any)
	v["is_request_finish"] = true
	return context.WithValue(ctx, piplineCtxValueKey, v)
}

func (p *PipLine) Result(ctxs ...context.Context) request.Response {
	if len(ctxs) > 0 {
		p.Ctx = ctxs[0]
	}
	ctx := p.Ctx
	insRes := make([][]byte, len(p.Ins))
	ctxmap := map[string]any{
		"current_request_index": 0,
		"is_request_finish":     false,
	}

	// process ins
	g, ctx := errgroup.WithContext(ctx)
	if p.Parall {

		for index, fn := range p.Ins {
			ctxmap["current_request_index"] = index
			ctx = context.WithValue(ctx, piplineCtxValueKey, ctxmap)
			fn := fn
			index := index
			g.Go(func() error {
				resp, err := fn(ctx, p.PipLineClient)
				if err != nil {
					return err
				}
				fmt.Println(string(resp))
				insRes[index] = resp
				return nil
			})
		}
		if err := g.Wait(); err != nil {
			return request.Response{Err: err}
		}
		ctxsetfinish(ctx)
	} else {
		for index, fn := range p.Ins {
			ctxmap["current_request_index"] = index
			ctx = context.WithValue(ctx, "values", ctxmap)
			resp, err := fn(ctx, p.PipLineClient)
			// fmt.Println(resp)
			if err != nil {
				return request.Response{Err: err}
			}
			insRes[index] = resp
			if index == len(p.Ins)-1 {
				ctxsetfinish(ctx)
			}
		}
		// fmt.Println(ctx.Value(piplineCtxValueKey))
	}
	// fmt.Println(ctx.Value(piplineCtxValueKey))
	// fmt.Println(insRes)
	// process out
	resp := p.Out(ctx, p.PipLineClient, insRes...)
	// fmt.Println(resp)
	return resp
}

func NewPipLine(opts ...PipLineOption) *PipLine {
	p := &PipLine{
		Ctx: context.Background(),
	}
	for _, opt := range opts {
		opt.apply(p)
	}
	return p
}
