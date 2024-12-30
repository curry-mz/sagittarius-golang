package finance

import (
	"context"
	"fmt"
	netHttp "net/http"

	"github.com/curry-mz/sagittarius-golang/app/proxy"
	"github.com/curry-mz/sagittarius-golang/cores/client/http"
	gErrors "github.com/curry-mz/sagittarius-golang/cores/errors"

	"github.com/pkg/errors"
)

func Query(ctx context.Context, req *QueryRequest) (map[int64]int64, error) {
	if len(req.PlayerIDs) == 0 {
		return nil, nil
	}
	c, err := proxy.InitHttpClient(ctx, Name)
	if err != nil {
		return nil, errors.WithMessage(err, "| proxy.InitHttpClient")
	}
	var rsp QueryResponse
	httpRsp, err := c.JsonPost(http.Request(ctx, queryUrl), req, &rsp)
	if err != nil {
		return nil, errors.WithMessage(err, "| c.JsonPost")
	}
	if httpRsp.StatusCode != netHttp.StatusOK {
		return nil, errors.New(fmt.Sprintf("post %s, httpCode:%d", queryUrl, httpRsp.StatusCode))
	}
	if rsp.Status != 0 {
		return nil, gErrors.New(rsp.Status, rsp.Message)
	}
	res := make(map[int64]int64)
	for _, d := range rsp.Data {
		res[d.PlayerID] = d.Chips
	}
	return res, nil
}

func Bet(ctx context.Context, req *BetRequest) error {
	c, err := proxy.InitHttpClient(ctx, Name)
	if err != nil {
		return errors.WithMessage(err, "| proxy.InitHttpClient")
	}
	var rsp BetResponse
	httpRsp, err := c.JsonPost(http.Request(ctx, betUrl), req, &rsp)
	if err != nil {
		return errors.WithMessage(err, "| c.JsonPost")
	}
	if httpRsp.StatusCode != netHttp.StatusOK {
		return errors.New(fmt.Sprintf("post %s, httpCode:%d", queryUrl, httpRsp.StatusCode))
	}
	if rsp.Status != 0 {
		return gErrors.New(rsp.Status, rsp.Message)
	}
	return nil
}

func Settle(ctx context.Context, req *SettleRequest) (*SettleData, error) {
	if len(req.Orders) == 0 {
		return nil, nil
	}
	c, err := proxy.InitHttpClient(ctx, Name)
	if err != nil {
		return nil, errors.WithMessage(err, "| proxy.InitHttpClient")
	}
	var rsp SettleResponse
	httpRsp, err := c.JsonPost(http.Request(ctx, settleUrl), req, &rsp)
	if err != nil {
		return nil, errors.WithMessage(err, "| c.JsonPost")
	}
	if httpRsp.StatusCode != netHttp.StatusOK {
		return nil, errors.New(fmt.Sprintf("post %s, httpCode:%d", settleUrl, httpRsp.StatusCode))
	}
	if rsp.Status != 0 {
		return nil, gErrors.New(rsp.Status, rsp.Message)
	}

	return &rsp.Data, nil
}
